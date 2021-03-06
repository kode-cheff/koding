KONFIG        = require 'koding-config-manager'
redis         = require 'redis'
REDIS_KEY     = 'social:disposable-email-addresses'
DOMAINS       = require 'disposable-email-domains'

LOW_QUALITY_DOMAINS = [
  '126.com'
  '139.com'
  '163.com'
  'example.com'
  'hushmail.com'
  'lavabit.com'
  'qq.com'
  'yahoo.com.ph'
  'yahoo.com.vn'

  # -- latest redis entries
  # -- ongoing discussion here: https://github.com/ivolo/disposable-email-domains/pull/22
  '10mail.org'
  '10minutemail.co.za'
  '123.com'
  '20mail.in'
  '20mail.it'
  'Upal.se'
  'abc.bg'
  'abusemail.de'
  'abv.bg'
  'abyssmail.com'
  'ac20mail.in'
  'adamsnet.info'
  'adm1.me'
  'agendosa.com'
  'antifork.org'
  'bladesmail.net'
  'bob.ubx.se'
  'boximail.com'
  'boxtemp.com.br'
  'byom.de'
  'camcecil.com'
  'ce.mintemail.com'
  'charklasers.com'
  'claeys.qc.to'
  'clrmail.com'
  'cock.li'
  'coloccini.com.ar'
  'computerquip.com'
  'crazespaces.pw'
  'cvs.in.th'
  'dadasoft.com.mx'
  'dani.ml'
  'dataleak.info.tm'
  'dauer.info'
  'digit-labs.web.id'
  'dinkys.ws'
  'divermail.com'
  'doc.biz.tm'
  'dodgemail.de'
  'dota.epicgamer.org'
  'dropmail.me'
  'eelmail.com'
  'enraged-bl.tk'
  'envenve1e.com'
  'epicgamer.org'
  'esa.thc.lv'
  'evilninjapirates.com'
  'faic.tk'
  'fakemail.com'
  'flurred.com'
  'fractum.hol.es'
  'gametheorylabs.com'
  'geer4.mooo.com'
  'gnoia.org'
  'grandmamail.com'
  'h1ch3r.net'
  'hacker1.com.br'
  'haqed.com'
  'haqued.com'
  'harakirimail.com'
  'hintz.org'
  'http://muell.email/'
  'humn.ws.gy'
  'idan.be'
  'ignorelist.com'
  'iiserk.net'
  'isaichkin.ru'
  'it.sackler.net'
  'kkll.cu.cc'
  'landmail.co'
  'lastmail.co'
  'lastmail.com'
  'lekovic.ca'
  'lichtisten.com'
  'linuxx.org'
  'linxlunx.info'
  'maildx.com'
  'mailed.ro'
  'mailfs.com'
  'mailme.crabdance.com'
  'mintemail.com'
  'misena.edu.co'
  'mohmal.com'
  'mouly.com.ar'
  'mrbox.root.sx'
  'muell.email'
  'my10minutemail.com'
  'mybox.root.sx'
  'national.shitposting.agency'
  'nikvdp.com'
  'novoemail.homenet.org'
  'nowonder.homenet.org'
  'nowwhat.linuxd.org'
  'onetime.email'
  'oos.tw'
  'ove.ali.com.pk'
  'p0ns.org'
  'paschke.org'
  'paukner.org'
  'pecdo.com'
  'pedco.com'
  'perlpowered.com'
  'plsmail.us.to'
  'protocultura.net'
  'pwn.linuxx.org'
  'qarea.com'
  'qasti.com'
  'qia.bep.co.id'
  'qisdo.com'
  'qisoa.com'
  'quasti.com'
  'radiku.ye.vc'
  'ranftl.org'
  'reddementes.net'
  'redthumb.info.tm'
  'rev.vci.si'
  'roshankarki.com.np'
  'samo.ohi.tw'
  'seco.sne.jp'
  'secret.shop.tm'
  'sempaktools.us.to'
  'sharklasers.com'
  'shokri.net'
  'shop.tm'
  'shop.tn'
  'sija.pl'
  'sixohquad.com'
  'sofasurfer.ch'
  'somany.ignorelist.com'
  'space-elephant.com'
  'spambooger.com'
  'spoofmail.de'
  'streetwisemail.com'
  'sysnet.org.pk'
  'tafmail.com'
  'techie.com'
  'tempmailer.com'
  'test5566.strangled.net'
  'thedens.org'
  'theins4ne.net'
  'throam.com'
  'tokenmail.de'
  'torry.multiservice.ru'
  'trash-mail.com'
  'travelwith.spacetechnology.net'
  'trbvm.com'
  'trickmail.net'
  'trickmail.net6'
  'trollope.za.net'
  'twilightparadox.com'
  'ubismail.net'
  'usa.com'
  'valemail.net'
  'viona.ml'
  'vomoto.com'
  'want.javafaq.nu'
  'welcome.twilightparadox.com'
  'wezel.ch'
  'wezel.info'
  'wickmail.net'
  'winter.org'
  'wollan.info'
  'wollmann.org'
  'xardas.eu'
  'xww.ro'
  'yert.ye.vc'
  'yet.eva.hk'
  'yeuthuong.org'
  'yis.vr.lt'
  'yomail.info'
  'you.loc.im'
]

DOMAINS = DOMAINS.concat LOW_QUALITY_DOMAINS

endsWith = (str, suffix) ->
  str.toLowerCase().indexOf(suffix.toLowerCase(), str.length - suffix.length) isnt -1

redisClient = null

check = (email) ->
  for domain in DOMAINS
    return no  if endsWith email, "@#{domain}"
  return yes

syncWithRedis = (callback) ->

  unless redisClient
    redisClient = redis.createClient(
      KONFIG.redis.port
      KONFIG.redis.host
      {}
    )

  redisClient.smembers REDIS_KEY, (err, domains) ->

    console.warn err  if err?
    domains ?= []

    DOMAINS.push domain for domain in domains when domain not in DOMAINS

    callback null


module.exports = emailchecker = (email, callback = -> ) ->

  unless email
    callback no
    return no

  syncWithRedis -> callback check email

  return check email
