package golib_ssh_test

import(
    ssh "github.com/weizhouBlue/golib_ssh"
    "fmt"
    "testing"
    "time"
)


func Test_pwd(t *testing.T){

    c:= &ssh.SshSession{
        ServerIPv4Ip: "10.6.185.10" ,
        Port : "11170" , 
    }

    if e:=c.ConnectByPwd("root" , "157ALp#!399") ; e!=nil {
        fmt.Printf("failed 1 , %v \n" , e )
        return
    }
    defer c.Close()

    if out, err , e:= c.ExecCmd("ls /" , "") ; e!=nil {
        fmt.Printf("failed 2 , %v \n" , e )
        fmt.Printf("stdout , %v \n" , err )
        return
    }else{
        fmt.Printf("out:  %v \n" , out )
        fmt.Printf("err:  %v \n" , err )

    }




    time.Sleep(20*time.Second)

    if out, err , e:= c.ExecCmd("ls /" , "" ) ; e!=nil {
        fmt.Printf("failed 2 , %v \n" , e )
        fmt.Printf("stderr , %v \n" , err )
        return
    }else{
        fmt.Printf("out:  %v \n" , out )
        fmt.Printf("err:  %v \n" , err )
    }

}




func Test_key(t *testing.T){

    c:= &ssh.SshSession{
        ServerIPv4Ip: "10.6.185.25" ,
        Port : "11170" , 
    }

    if e:=c.ConnectByPublicKey("root" , "/Users/weizhoulan/.ssh/id_rsa") ; e!=nil {
        fmt.Printf("failed 1 , %v \n" , e )
        return
    }
    defer c.Close()


    if out, err , e:= c.ExecCmd("ls /" , "" ) ; e!=nil {
        fmt.Printf("failed 2 , %v \n" , e )
        fmt.Printf("stderr , %v \n" , err )

        return
    }else{
        fmt.Printf("out:  %v \n" , out )
        fmt.Printf("err:  %v \n" , err )
    }

    time.Sleep(10*time.Second)

    if out, err , e:= c.ExecCmd("ls /" , "" ) ; e!=nil {
        fmt.Printf("failed 2 , %v \n" , e )
        fmt.Printf("stderr , %v \n" , err )
        return
    }else{
        fmt.Printf("out:  %v \n" , out )
        fmt.Printf("err:  %v \n" , err )
    }

}










