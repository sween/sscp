# InterSystems Superserver Copy
File Transfer over the InterSystems IRIS SuperServer.

## Why
Dropping to $ZF, the shell, or automation tooling like Ansible is proving to cause more problems sometimes than its worth when automating systems for post configuration for things like MIRRORING setup, certificates, and libraries across systems.  As a common practice, it would be nice to rule out "ObjectScript Native" mechanisms before resorting to doing those things outside the instance... this solution allows you to rule it out for file transfer and keep the file copy operations across instances "ObjectScript Native".

<img src="https://github.com/sween/sscp/raw/main/assets/sscp_copy.gif" alt="Example Copy">

## Setup
You need install the ipm package on all instances and in selected namespaces you wish to connect to. The namespace is not directly important to enabling the file copy operation, but just note it does not have to be in %SYS to work.

```
zn <Your Favorite Namespace>  
zpm "install sscp"  
```

## Usage
For the initial version, a full connection string is required, and is in the form of `username:password@host:port@namespace:file`.

So, to transfer an IRIS.DAT file from host `fhirwatch-pop-os` (laptop) to remote server `target-host` you could use the following:

```
Set tSC = ##class(ZMSP.SuperServer).Copy("_SYSTEM:SYS@fhirwatch-pop-os:1972@%SYS:/tmp/source.dat","kmitnick:12345@target-host:1972@SDOH:/tmp/destination.dat")
```
<img src="https://github.com/sween/sscp/raw/main/assets/sscp_server2server.png" alt="Server to Server Transfer">

Do I have to be on the instances I am transferring files across?  Nope.

Let's say you've loaded the ipm package on your laptop, container, or toaster running IRIS, and you have two targets you want to do a file transfer across:

```
Set tSC = ##class(ZMSP.SuperServer).Copy("_SYSTEM:SYS@remote-host:1972@%SYS:/tmp/source.dat","kmitnick:12345@target-host:1972@SDOH:/tmp/destination.dat")
```
<img src="https://github.com/sween/sscp/raw/main/assets/sscp_thirdpartyinitiator.png" alt="Third Party Initiator">

## Use Cases

Establishing HealthShare Mirroring with HSSYS.  
Automated provisioning of MIRRORS.  
Transfer and update of Globals or Classes for those things outside of mirroring (%SYS).  
Getting Files into your Health Connect Cloud instance to the file system interoperability target `/connect`.  

## Configuring Auditing
One of the benefits here of keeping this operation in house is it is now auditable, to enable auditing you can create a source and log the operation. Programmatically:

```objectscript
zn "%SYS"
set sc = ##class(Security.Events).Create("sscp","File Transfer", "ConfigurationChange","File Transfer Initiated",1)
```

Through the UI:

1. Go to Management Portal -> System Administration -> Security -> Auditing -> Configure User Events
2. Press button Create New Event
3. Set Event Source: sscp
4. Set Event Type: File Transfer
5. Set Event Name: SSCP
6. Press Save

<img src="https://github.com/sween/sscp/raw/main/assets/sscp_auditevent.png" alt="Create Audit Event">

Then, after your $$$OK, you can do:

```
Set tSC = $SYSTEM.Security.Audit("sscp","File Transfer", "ConfigurationChange","Transferred IRIS.DAT", "Setup of HealthShare Mirroring")
```

Which results in something like:

<img src="https://github.com/sween/sscp/raw/main/assets/sscp_loggedauditevent.png" alt="Logged Audit Event">

Credit:
[@Yuri.Gomes - Leveraging Audit Database](https://community.intersystems.com/post/leveraging-audit-database)


## Is it fast?

Yeah I dunno... Most of the transfers in production use are under 100MB at the moment (HSSYS out of the box is around 22MB), but the only other benchmark I can share is some stuff that was around 800MB and took about a minute and a half traversing the internet.

## Troubleshooting

The app logs to messages.log on both sides loosely.

## Definitive RoadMap

- [ ] Use stored connections instead to avoid secrets all over the place and ridiculous connection strings.
- [ ] Investigate impact of the use private SSL Certificates wuth the connections.

## 

## Authors

[@Ron.Sweeney1582](https://community.intersystems.com/user/sween-sweeney)  
[@Eduard.Lebedyuk](https://community.intersystems.com/user/eduard-lebedyuk)  

