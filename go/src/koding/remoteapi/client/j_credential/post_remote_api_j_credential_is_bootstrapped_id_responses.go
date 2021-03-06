package j_credential

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	"koding/remoteapi/models"
)

// PostRemoteAPIJCredentialIsBootstrappedIDReader is a Reader for the PostRemoteAPIJCredentialIsBootstrappedID structure.
type PostRemoteAPIJCredentialIsBootstrappedIDReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PostRemoteAPIJCredentialIsBootstrappedIDReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewPostRemoteAPIJCredentialIsBootstrappedIDOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewPostRemoteAPIJCredentialIsBootstrappedIDOK creates a PostRemoteAPIJCredentialIsBootstrappedIDOK with default headers values
func NewPostRemoteAPIJCredentialIsBootstrappedIDOK() *PostRemoteAPIJCredentialIsBootstrappedIDOK {
	return &PostRemoteAPIJCredentialIsBootstrappedIDOK{}
}

/*PostRemoteAPIJCredentialIsBootstrappedIDOK handles this case with default header values.

OK
*/
type PostRemoteAPIJCredentialIsBootstrappedIDOK struct {
	Payload *models.JCredential
}

func (o *PostRemoteAPIJCredentialIsBootstrappedIDOK) Error() string {
	return fmt.Sprintf("[POST /remote.api/JCredential.isBootstrapped/{id}][%d] postRemoteApiJCredentialIsBootstrappedIdOK  %+v", 200, o.Payload)
}

func (o *PostRemoteAPIJCredentialIsBootstrappedIDOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.JCredential)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
