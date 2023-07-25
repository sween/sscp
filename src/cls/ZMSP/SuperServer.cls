Class ZMSP.SuperServer
{
    // username:password@host:port@namespace:file
    ClassMethod Copy(pSource As %String, pTarget As %String) As %Status
    {
        Set tSC = $$$OK
        Try {

            // Check Supplied Parameters
            Set tSC = ..CheckConnectionString(pSource)
            Set tSC = ..CheckConnectionString(pTarget)

            // Check Connections %Net.DB.Iris
            Set tSC = ..BuildConnection(pSource, .tSourceConn)
            Set tSC = ..BuildConnection(pTarget, .tDestConn)

            // Get files
            Set tSourceNamespaceFile = $PIECE(pSource, "@", 3)
            Set tTargetNamespaceFile = $PIECE(pTarget, "@", 3)
            Set tSourceFile = $PIECE(tSourceNamespaceFile, ":", 2)
            Set tTargetFile = $PIECE(tTargetNamespaceFile, ":", 2)


            // Initiate Transfer
            Set tSC = ..InitiateTransfer(tSourceConn, pTarget, tSourceFile, tTargetFile)


        } Catch {

            Throw $$$ERROR($$$GeneralError, "Something went wrong Somewhere, check messages.log.")

        }
         

        Quit tSC
    }

    // username:password@host:port@namespace:file
    ClassMethod CheckConnectionString(pConnection As %String, ByRef pConn As %Net.DB.Iris) As %Status
    {
        Set tSC = $$$OK
        Try {
            Set tUserPass = $PIECE(pConnection, "@", 1)
            Set tHostPort = $PIECE(pConnection, "@", 2)
            Set tNamespaceFile = $PIECE(pConnection, "@", 3)

            Set tUserName = $PIECE(tUserPass, ":", 1)
            Set tPassword = $PIECE(tUserPass, ":", 2)
            
            Set tHost = $PIECE(tHostPort, ":", 1)
            Set tPort = $PIECE(tHostPort, ":", 2)
            
            Set tNameSpace = $PIECE(tNamespaceFile, ":", 1)
            Set tFile = $PIECE(tNamespaceFile, ":", 2)

            if tUserName,tPassword,tHost,tPort,tNameSpace,tFile = "" {
                Throw $$$ERROR($$$GeneralError, "Supplied Connection string would not parse.")
            }
        }
        Catch {
            Do ##class(%SYS.System).WriteToConsoleLog("Supplied Connection string would not parse.")
            Throw $$$ERROR($$$GeneralError, "Supplied Connection string would not parse.")
        }

        Quit tSC
    }

    // %Net.DB.Iris
    ClassMethod BuildConnection(pConnection As %String, ByRef pConnection As %Net.DB.Iris) As %Status
    {
        Set tSC = $$$OK
        Try {
            
            Set tUserPass = $PIECE(pConnection, "@", 1)
            Set tHostPort = $PIECE(pConnection, "@", 2)
            Set tNamespaceFile = $PIECE(pConnection, "@", 3)

            Set tUserName = $PIECE(tUserPass, ":", 1)
            Set tPassword = $PIECE(tUserPass, ":", 2)
            
            Set tHost = $PIECE(tHostPort, ":", 1)
            Set tPort = $PIECE(tHostPort, ":", 2)
            
            Set tNameSpace = $PIECE(tNamespaceFile, ":", 1)
            
            set tConnection = ##class(%Net.DB.DataSource).CreateConnection(tHost, tPort, tNameSpace, tUserName, tPassword)
            set tIris = tConnection.CreateIris()
            Set pConnection = tIris
            
		} Catch {
            Do ##class(%SYS.System).WriteToConsoleLog("SuperServer Copy connection could not be created or established.")
			Throw $$$ERROR($$$GeneralError, "SuperServer Copy connection could not be created or established.")
		}

        Quit tSC
    }

    // FromSource execute Transfer on Source with a Target of Target (mirrora)
    ClassMethod InitiateTransfer(pSourceConnection As %Net.DB.Iris, pTarget As %String, pSourceFile As %String, pTargetFile As %String) As %Status
    {
        Set tSC = $$$OK
        // Run "Transfer" on "Source" with "Target"
        Try {
            Set tSC = pSourceConnection.ClassMethodValue("ZMSP.SuperServer", "Transfer", pTarget, pSourceFile, pTargetFile)
        }
        Catch {
            Do ##class(%SYS.System).WriteToConsoleLog("SuperServer Copy connection could not be initiated.")
			Throw $$$ERROR($$$GeneralError, "SuperServer Copy connection could not be initiated.")
        }

        Quit tSC
    }

    // to mirrorb
    ClassMethod Transfer(pTarget As %String, pSourceFile As %String, pTargetFile As %String) As %Status
    {
        Set tSC = $$$OK

        Try {
            Set tSC = ..BuildConnection(pTarget, .pDestConn)

            set stream = ##class(%Stream.FileBinary).%New()
            do stream.LinkToFile(pSourceFile)
            while 'stream.AtEnd {
                set chunk = stream.Read($$$MaxStringLength/3)
                do pDestConn.ClassMethodVoid("ZMSP.Mirror", "WriteBinary", pTargetFile, chunk)
            }

        }
        Catch {
            Do ##class(%SYS.System).WriteToConsoleLog("SuperServer Copy failure in Transfer")
			Throw $$$ERROR($$$GeneralError, "SuperServer Copy failure in Transfer.")

        }

        Quit tSC
    }

    /*
    /// Target Connection to Target
    /// ("mirrora","/data/sys/hssys/mgr/MIRROR.DAT", "/data/sys/hssys/mgr/MIRROR.DAT")
    ClassMethod Transfer(host, hostfile, localfile) As %Status
    {
        Set tSC = $$$OK
        //set s=##class(%Stream.FileBinary).%New()
        set host=host,port=1972,namespace="%SYS",user="intersystems"
        set connection = ##class(%Net.DB.DataSource).CreateConnection(host, port, namespace, user, user)
        set iris = connection.CreateIris()
        
        set stream = ##class(%Stream.FileBinary).%New()
        do stream.LinkToFile(hostfile)
        while 'stream.AtEnd {
            set chunk = stream.Read($$$MaxStringLength/3)
            do iris.ClassMethodVoid("ZMSP.Mirror", "WriteBinary", localfile, chunk)
        }
        Quit tSC
    }
    */

    ClassMethod WriteBinary(pFile, pString)
    {
        set stream = ##class(%Stream.FileBinary).%New()
        do stream.LinkToFile(pFile)
        do stream.MoveToEnd()
        do stream.Write(pString)
        do stream.%Save()
        kill stream
    }

}