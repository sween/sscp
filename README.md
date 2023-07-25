# InterSystems Superserver Copy
File Transfer over the InterSystems IRIS Superserver.

## Use Case
Dropping to $ZF is proving to cause more problems when automating systems.

## Setup

```
zn "%SYS"
zpm "install sscp"
```

## Usage

```
Set tSC = ##class(ZMSP.SuperServer).Copy("_SYSTEM:SYS@fhirwatch-pop-os:1972@%SYS:/tmp/source.dat","kmitnick:12345@druidiaairshield.com:1972@%SYS:/tmp/destination.dat")
```

## Definitive RoadMap
Use stored connections instead to avoid secrets all over the place and ridiculous connection strings.
Investigate private SSL Certificates.

## Authors

Eduard
Ron