Class ZMSP.SuperServer
{
    // Source and Target are in the format of username:password@host:port@namespace:file
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

    // This is just a filter to see if the connection strings can be parsed.
    ClassMethod CheckConnectionString(pConnection As %String) As %Status
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

    // Check to see if a connection can be made and returns a %Net.DB.Iris
    ClassMethod BuildConnection(pConnection As %String, ByRef pConn As %Net.DB.Iris) As %Status
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
            Set pConn = tIris
            
		} Catch {
            Do ##class(%SYS.System).WriteToConsoleLog("SuperServer Copy connection could not be created or established.")
			Throw $$$ERROR($$$GeneralError, "SuperServer Copy connection could not be created or established.")
		}

        Quit tSC
    }

    // This method connects to the source and grabs the paylod of the source file
    ClassMethod InitiateTransfer(pSourceConnection As %Net.DB.Iris, pTarget As %String, pSourceFile As %String, pTargetFile As %String) As %Status
    {
        Set tSC = $$$OK

        Try {
            Set tSC = pSourceConnection.ClassMethodValue("ZMSP.SuperServer", "Transfer", pTarget, pSourceFile, pTargetFile)
        }
        Catch {
            Do ##class(%SYS.System).WriteToConsoleLog("SuperServer Copy connection could not be initiated.")
			Throw $$$ERROR($$$GeneralError, "SuperServer Copy connection could not be initiated.")
        }

        Quit tSC
    }

    // This method is invoked on the source per InitiateTransfer, and "pushes" the file to the Target
    ClassMethod Transfer(pTarget As %String, pSourceFile As %String, pTargetFile As %String) As %Status
    {
        Set tSC = $$$OK

        Try {
            Set tSC = ..BuildConnection(pTarget, .pDestConn)
            set stream = ##class(%Stream.FileBinary).%New()
            do stream.LinkToFile(pSourceFile)
            while 'stream.AtEnd {
                // Adjust this if you want better performance.
                set chunk = stream.Read($$$MaxStringLength/3)
                do pDestConn.ClassMethodVoid("ZMSP.SuperServer", "WriteBinary", pTargetFile, chunk)
            }

        }
        Catch {
            Do ##class(%SYS.System).WriteToConsoleLog("SuperServer Copy failure in Transfer")
			Throw $$$ERROR($$$GeneralError, "SuperServer Copy failure in Transfer.")

        }

        Quit tSC
    }


    ClassMethod WriteBinary(pFile As %String, pString As %String)
    {
        set stream = ##class(%Stream.FileBinary).%New()
        do stream.LinkToFile(pFile)
        do stream.MoveToEnd()
        do stream.Write(pString)
        do stream.%Save()
        kill stream
    }

}