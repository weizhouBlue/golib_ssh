package golib_ssh
import (
	"golang.org/x/crypto/ssh"
	"fmt"
	"strconv"
	"net"
	"io/ioutil"
	"bytes"
	"strings"
	"os"
	"runtime"
)

//==================================
/*
https://godoc.org/golang.org/x/crypto/ssh
*/



//============================================

var (
    EnableLog=false	
)



//====================log==========================

func getFileName( path string ) string {
    b:=strings.LastIndex(path,"/")
    if b>=0 {
        return path[b+1:]
    }else{
        return path
    }
}

func log( format string, a ...interface{} ) (n int, err error) {

    if EnableLog {

		prefix := ""
	    funcName,filepath ,line,ok := runtime.Caller(1)
	    if ok {
	    	file:=getFileName(filepath)
	    	funcname:=getFileName(runtime.FuncForPC(funcName).Name())
	    	prefix += "[" + file + " " + funcname + " " + strconv.Itoa(line) +  "]     "
	    }

        return fmt.Printf(prefix+format , a... )    
    }
    return  0,nil
}





//===================================

type SshSession struct {
	ServerIPv4Ip string
	Port string
	sshClient *ssh.Client

}



func CheckIPv4Format( ip string ) bool {
	result := net.ParseIP(ip)
	if result==nil {
		return false
	}
	if result.To4()==nil {
		return false
	}
	return true
}


func (s *SshSession ) checkConfig() error {
	if len( s.ServerIPv4Ip )==0 {
		return fmt.Errorf("ServerIPv4Ip is empty")
	}
	if CheckIPv4Format(s.ServerIPv4Ip)==false{
		return fmt.Errorf("ServerIPv4Ip is not an ipv4 address , %v ", s.ServerIPv4Ip )		
	}

	if len( s.Port )==0 {
		return fmt.Errorf("Port is empty")
	}


	if n , err := strconv.ParseInt( s.Port , 10, 64); err == nil {
	    if n<=0 || n>=65536 {
			return fmt.Errorf("Port is out of range 1-65536 , %v " , s.Port  )
	    }
	}else{
		return fmt.Errorf("Port is number , %v" , s.Port)
	}

	return nil

}


func ( s *SshSession ) ConnectByPwd( userName , passwd string ) ( err error) {

	log("ConnectByPwd to server %v:%v with userName=%v , passwd=%v " ,s.ServerIPv4Ip, s.Port , userName , passwd  )

	if e:= s.checkConfig() ; e!=nil {
		return e
	}

	if len(userName)==0 {
		return fmt.Errorf("userName is empty , %v")
	}
	if len(passwd)==0 {
		return fmt.Errorf("passwd is empty , %v")
	}



	// https://godoc.org/golang.org/x/crypto/ssh#ClientConfig
	config := &ssh.ClientConfig{
	    User: userName ,
	    // 认证方式
	    Auth: []ssh.AuthMethod{
	        ssh.Password( passwd ),
	    },
	    // HostKeyCallback is called during the cryptographic handshake to validate the server's host key
	    //HostKeyCallback: ssh.FixedHostKey(hostKey),
	    HostKeyCallback: ssh.InsecureIgnoreHostKey(),

	}

	// https://godoc.org/golang.org/x/crypto/ssh#Dial
	s.sshClient , err = ssh.Dial("tcp", s.ServerIPv4Ip+":"+s.Port , config)
	if err != nil {
	    return err
	}

	return nil

}




func ( s *SshSession ) ConnectByPublicKey( userName , privateKeyPath string ) (err  error) {

	log("ConnectByPublicKey to server %v:%v with userName=%v , privateKeyPath=%v " ,s.ServerIPv4Ip, s.Port , userName , privateKeyPath  )

	if e:= s.checkConfig() ; e!=nil {
		return e
	}

	if len(userName)==0 {
		return fmt.Errorf("userName is empty ")
	}
	if len(privateKeyPath)==0 {
		return fmt.Errorf("privateKeyPath is empty ")
	}

	if _, e := os.Stat(privateKeyPath) ; e != nil {
		return fmt.Errorf("privateKeyPath is wrong , %v" , e )
	}



	// A public key may be used to authenticate against the remote
	// server by using an unencrypted PEM-encoded private key file.
	//
	// If you have an encrypted private key, the crypto/x509 package
	// can be used to decrypt it.
	//  "/home/user/.ssh/id_rsa" 
	key, err := ioutil.ReadFile( privateKeyPath )
	if err != nil {
		return fmt.Errorf("failed to ready private key=%v , info=%v" ,privateKeyPath , err )
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("failed to parse private key=%v , info=%v" ,privateKeyPath , err )
	}


	// https://godoc.org/golang.org/x/crypto/ssh#ClientConfig
	config := &ssh.ClientConfig{
	    User: userName ,
	    Auth: []ssh.AuthMethod{
	        // Use the PublicKeys method for remote authentication.
	        ssh.PublicKeys(signer),
	    },
	    //HostKeyCallback: ssh.FixedHostKey(hostKey),
	    HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}


	// https://godoc.org/golang.org/x/crypto/ssh#Dial
	s.sshClient , err = ssh.Dial("tcp", s.ServerIPv4Ip+":"+s.Port , config)
	if err != nil {
	    return err
	}

	return nil

}



/*
cmd: 运行命令
stdin: 命令的标准输入

如果命令返回码0，则err=nil , 否则 err 报错
*/
func ( s *SshSession ) ExecCmd( cmd string , stdin string ) (  stdout , stderr string , err error ){

	log("ExecCmd to server %v:%v with cmd=%v , stdin=%v " ,s.ServerIPv4Ip, s.Port , cmd , stdin  )

	if s.sshClient==nil {
		return  "" ,   "" , fmt.Errorf("session is not still set up " )
	}


	session, err := s.sshClient.NewSession()
	if err != nil {
		return "" ,  "" , fmt.Errorf("failed to create session for cmd=%v , info=%v " , cmd ,  err )
	}
	defer session.Close()


	// 开启 TTY, 虽然是没必要的，但是，如果不开，发现 server上会有一些 USERNAME@notty 的进程产生
	// modes := ssh.TerminalModes{
	//     ssh.ECHO:     0,   // disable echoing
	//     ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
	//     ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	// }
	// if e := session.RequestPty("xterm", 80, 40, modes); e != nil {
	//     return "", "" , fmt.Errorf("failed to create tty, info=%v " ,  err )
	// }


	var m , n  bytes.Buffer

	session.Stdout = &m
	session.Stderr = &n
	if len(stdin)!=0 {
		session.Stdin=strings.NewReader(stdin)
	}

	if err := session.Run(cmd) ; err != nil {
		return m.String() ,   n.String() , fmt.Errorf("failed to run command=%v , info=%v ,stderr=%v " , cmd,  err , n.String() )
	}

	return m.String() , n.String()  , nil 

}


func ( s *SshSession ) Close(  )  error {

	return s.sshClient.Close()

}
