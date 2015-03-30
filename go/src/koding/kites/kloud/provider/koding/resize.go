package koding

import (
	"fmt"
	"koding/kites/kloud/machinestate"
	"strconv"
	"time"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"github.com/mitchellh/goamz/ec2"
	"golang.org/x/net/context"
)

// Resize increases the current machines underling volume to a larger volume
// without affecting or destroying the users data.
func (m *Machine) Resize(ctx context.Context) (resErr error) {
	// Please read the steps before you dig into the code and try to change or
	// fix something. Intented lines are cleanup or self healing procedures
	// which should be called in a defer - arslan:
	//
	// 0. Check if size is eligible (not equal or less than the current size)
	// 1. Prepare/Get volumeId of current instance
	// 2. Prepare/Get availabilityZone of current instance
	// 3. Stop the instance so we can get the snapshot
	// 4. Create new snapshot from the current volumeId of that stopped instance
	//		4a. Delete snapshot after we are done with all following steps (not needed anymore)
	// 5. Create new volume with the desired size from the snapshot and same zone.
	//		5a. Delete volume if something goes wrong in following steps
	// 6. Detach the volume of current stopped instance.
	//    if something goes wrong:
	//		6a. Detach new volume, attach old volume. New volume will be
	//		    attached in the following step, so we are going to rewind it.
	//	  if everything is ok:
	//		6b. Delete old volume (not needed anymore)
	// 7. Attach new volume to current stopped instance
	// 8. Start the stopped instance with the new larger volume
	// 9. Update Default Domain record with the new IP (stopping/starting changes the IP)
	// 11 Update Domain aliases with the new IP (stopping/starting changes the IP)
	// 12. Check if Klient is running

	if err := m.UpdateState("Machine is resizing", machinestate.Pending); err != nil {
		return err
	}

	m.push("Resizing initialized", 10, machinestate.Pending)

	a := m.Session.AWSClient

	m.Log.Info("checking if size is eligible for instance %s", a.Id())
	instance, err := a.Instance()
	if err != nil {
		return err
	}

	if len(instance.BlockDevices) == 0 {
		return fmt.Errorf("fatal error: no block device available")
	}

	// we need it in a lot of places!
	oldVolumeId := instance.BlockDevices[0].VolumeId

	oldVolResp, err := a.Client.Volumes([]string{oldVolumeId}, ec2.NewFilter())
	if err != nil {
		return err
	}

	volSize := oldVolResp.Volumes[0].Size
	currentSize, err := strconv.Atoi(volSize)
	if err != nil {
		return err
	}

	desiredSize := a.Builder.StorageSize

	m.Log.Debug("DesiredSize: %d, Currentsize %d", desiredSize, currentSize)

	// Storage is counting all current sizes. So we need ask only for the
	// difference that we want to add. So say if the current size is 3
	// and our desired size is 10, we need to ask if we have still
	// limit for a 7 GB space.
	if err := m.Checker.Storage(desiredSize-currentSize, m.Username); err != nil {
		return err
	}

	m.push("Checking if size is eligible", 20, machinestate.Pending)

	m.Log.Info("user wants size '%dGB'. current storage size: '%dGB'", desiredSize, currentSize)
	if desiredSize <= currentSize {
		return fmt.Errorf("resizing is not allowed. Desired size: %dGB should be larger than current size: %dGB",
			desiredSize, currentSize)
	}

	if desiredSize > 100 {
		return fmt.Errorf("resizing is not allowed. Desired size: %d can't be larger than 100GB",
			desiredSize)
	}

	m.push("Stopping old instance", 30, machinestate.Pending)
	m.Log.Info("stopping instance %s", a.Id())
	if m.State() != machinestate.Stopped {
		err := m.Session.AWSClient.Stop(ctx)
		if err != nil {
			return err
		}
	}

	m.push("Creating new snapshot", 40, machinestate.Pending)
	m.Log.Info("creating new snapshot from volume id %s", oldVolumeId)
	snapshotDesc := fmt.Sprintf("Temporary snapshot for instance %s", instance.InstanceId)
	snapshot, err := a.CreateSnapshot(oldVolumeId, snapshotDesc)
	if err != nil {
		return err
	}
	newSnapshotId := snapshot.Id

	defer func() {
		m.Log.Info("deleting snapshot %s (not needed anymore)", newSnapshotId)
		a.DeleteSnapshot(newSnapshotId)
	}()

	m.push("Creating new volume", 50, machinestate.Pending)
	m.Log.Info("creating volume from snapshot id %s with size: %d", newSnapshotId, desiredSize)

	// Go on with the current volume type. SSD(gp2) or Magnetic(standard)
	volType := oldVolResp.Volumes[0].VolumeType

	volume, err := a.CreateVolume(newSnapshotId, instance.AvailZone, volType, desiredSize)
	if err != nil {
		return err
	}
	newVolumeId := volume.VolumeId

	// delete volume if something goes wrong in following steps
	defer func() {
		if resErr != nil {
			m.Log.Info("(an error occurred) deleting new volume %s ", newVolumeId)
			_, err := a.Client.DeleteVolume(newVolumeId)
			if err != nil {
				m.Log.Error(err.Error())
			}
		}
	}()

	m.push("Detaching old volume", 60, machinestate.Pending)
	m.Log.Info("detaching current volume id %s", oldVolumeId)
	if err := a.DetachVolume(oldVolumeId); err != nil {
		return err
	}

	// reattach old volume if something goes wrong, if not delete it
	defer func() {
		// if something goes wrong  detach the newly attached volume and attach
		// back the old volume  so it can be used again
		if resErr != nil {
			m.Log.Info("(an error occurred) detaching newly created volume volume %s ", newVolumeId)
			if err := a.DetachVolume(newVolumeId); err != nil {
				m.Log.Error("couldn't detach: %s", err.Error())
			}

			m.Log.Info("(an error occurred) attaching back old volume %s", oldVolumeId)
			if err = a.AttachVolume(oldVolumeId, a.Id(), "/dev/sda1"); err != nil {
				m.Log.Error("couldn't attach: %s", err.Error())
			}
		} else {
			// if not just delete, it's not used anymore
			m.Log.Info("deleting old volume %s (not needed anymore)", oldVolumeId)
			go a.Client.DeleteVolume(oldVolumeId)
		}
	}()

	m.push("Attaching new volume", 70, machinestate.Pending)
	// attach new volume to current stopped instance
	if err := a.AttachVolume(newVolumeId, a.Id(), "/dev/sda1"); err != nil {
		return err
	}

	m.push("Starting instance", 80, machinestate.Pending)
	// start the stopped instance now as we attached the new volume
	instance, err = m.Session.AWSClient.Start(ctx)
	if err != nil {
		return err
	}
	m.IpAddress = instance.PublicIpAddress

	m.push("Updating domain", 85, machinestate.Pending)

	if err := m.Session.DNSClient.Validate(m.Domain, m.Username); err != nil {
		m.Log.Error("couldn't update machine domain: %s", err.Error())
	}
	if err := m.Session.DNSClient.Upsert(m.Domain, m.IpAddress); err != nil {
		m.Log.Error("couldn't update machine domain: %s", err.Error())
	}

	m.push("Updating domain aliases", 87, machinestate.Pending)
	// also get all domain aliases that belongs to this machine and unset
	domains, err := m.Session.DNSStorage.GetByMachine(m.Id.Hex())
	if err != nil {
		m.Log.Error("fetching domains for unsetting err: %s", err.Error())
	}

	for _, domain := range domains {
		if err := m.Session.DNSClient.Validate(domain.Name, m.Username); err != nil {
			m.Log.Error("couldn't update machine domain: %s", err.Error())
		}
		if err := m.Session.DNSClient.Upsert(domain.Name, m.IpAddress); err != nil {
			m.Log.Error("couldn't update machine domain: %s", err.Error())
		}
	}

	m.push("Checking remote machine", 90, machinestate.Pending)
	m.Log.Info("connecting to remote Klient instance")
	m.checkKite()

	return m.Session.DB.Run("jMachines", func(c *mgo.Collection) error {
		return c.UpdateId(
			m.Id,
			bson.M{"$set": bson.M{
				"ipAddress":         m.IpAddress,
				"meta.instanceName": m.Meta.InstanceName,
				"meta.instanceId":   m.Meta.InstanceId,
				"meta.instanceType": m.Meta.InstanceType,
				"status.state":      machinestate.Running.String(),
				"status.modifiedAt": time.Now().UTC(),
				"status.reason":     "Machine is running",
			}},
		)
	})

	return nil
}
