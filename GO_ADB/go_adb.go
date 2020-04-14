package GO_ADB

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	//"io"
	"syscall"
	"bytes"
	"time"
	"path"
)


type Adb struct{
	SerialNumber string

}

func NewAdb(SerialNumber string) *Adb{
	return &Adb{SerialNumber}
}

func (this *Adb)Push(src,dst string) int{
	return 0
}

func (this *Adb)Pull(src,dst string) int{
	
	conn:=this.adb_connect([]byte("sync:"))
	if conn==nil{
		fmt.Fprintf(os.Stderr,"Pull: establish communication failed.")
		return -1
	}
	buf1:=make([]byte,512)
	len,err:=conn.Read(buf1)
	fmt.Printf("sync: %d:%s",len,buf1)
	if err!=nil{
		fmt.Fprintf(os.Stderr,"Pull: read from server failed, err: %s\n",err)
		return -1
	}
	var mode uint32
	if sync_readmode(conn,src,&mode)!=0{
		fmt.Fprintf(os.Stderr,"Pull:sync_readmode failed\n")
		return -1
	}
	if mode==0{
		fmt.Fprintf(os.Stderr,"remote object '%s' does not exist\n",src)
	}
	if ISREG(mode) || ISLINK(mode) || ISCHR(mode) || ISBLK(mode) {
		fileinfo,err:=os.Stat(dst)
		if err==nil{
			if fileinfo.IsDir(){
				filename:=DirStops(src)
				dst=path.Join(dst,filename)
				fmt.Println(dst)
			}
		}
	}

	//begin sync
	Sync_recv(conn,src,dst)
	return 0
}

func (this *Adb)Kill_server() int {
	conn:=this.adb_connect([]byte("host:kill"))
	if conn==nil{
		fmt.Fprintf(os.Stderr,"Kill_server:kill server failed.")
		return -1
	}
	return 0
}

/* create adb interact shell*/
func (this *Adb)Interactive_shell()int{
	
	conn:=this.adb_connect([]byte("shell:"))
	if conn==nil{
		fmt.Println("failed to make shell")
		return 1
	}
	go read(conn)
	writeToShell(conn)
	//for conn!=nil{}
	return 0
}
func switch_socket_transport(conn net.Conn,sn string) int{
	//fmt.Println("switch_socket_transport")
	trimmedInput:=fmt.Sprintf("host:transport:%s",sn)
	trimmedInput=fmt.Sprintf("%04x%s",strings.Count(trimmedInput,"")-1,trimmedInput)
	_,err:=conn.Write([]byte(trimmedInput))
	if err!=nil{
		fmt.Println("switch socket transport failed:",err)
		conn.Close()
		return -1
	}
	return 0

}
func (this *Adb) adb_connect(c []byte) net.Conn{
	conn:=this._adb_connect([]byte("host:version"))
	if conn==nil{
		fmt.Println("failed to start daemon")
	}else{
		buf:=make([]byte,16)
		conn.Read(buf)
		//fmt.Println("adb version:",string(buf[8:]))
		conn.Close()
	}
	conn=this._adb_connect(c)
	
	if conn==nil{
		fmt.Println("failed to start service")
		return nil
	}
	return conn
}
func (this *Adb) _adb_connect(c []byte) net.Conn{
	
	len:=bytes.Count(c,nil)-1
	if len<1 || len>1024  {
		fmt.Println("cmd too long")
		return nil
	}
	trimmedInput:=fmt.Sprintf("%04x%s",len,c)
	//fmt.Println("trimmedInput: ", trimmedInput,string(c[:3]))
	conn:=create_socket()
	if conn==nil{
		return nil
	}
	if !bytes.Equal(c[:4],[]byte("host")) && switch_socket_transport(conn,this.SerialNumber)!=0 {
		
		return nil
		
		
	}
	//fmt.Println("Write cmd  ")
	time.Sleep(1)
	_,err:=conn.Write([]byte(trimmedInput))
	if err!=nil{
		fmt.Printf("Write cmd %s failed:%s \n",trimmedInput, err)
		conn.Close()
		return nil
	}
	//可加入判断server是否回复OKAY or FAIL ,TODO
	return conn

	
}


func create_socket()net.Conn{
	conn, err := net.Dial("tcp", "localhost:5037")
	//defer conn.Close()
	if err != nil {
		//由于目标计算机积极拒绝而无法创建连接
		fmt.Println("Error dialing", err.Error())
		return nil // 终止程序
	}
	return conn
}

func writeToShell(conn net.Conn){
	inputReader := bufio.NewReader(os.Stdin)

	// 给服务器发送信息直到程序退出
	for conn!=nil {
		input, _ := inputReader.ReadString('\n')
		trimmedInput := strings.Trim(input, "\r\n")
		if trimmedInput == "Q" {
			return
		}
		time.Sleep(1)
		_, err := conn.Write([]byte(trimmedInput+"\n"))
		if err !=nil{
			//fmt.Printf("write shell command failed\n")
			break
		}
	}
}
func read(conn net.Conn) {
	buf:=make([]byte,4096)
	for conn!=nil{
		len,err:=conn.Read(buf)
		//fmt.Println("line %s",buf,len)
		time.Sleep(1)
		fmt.Printf(string(buf[:len]))
		if len==0{
			//fmt.Println("disconnect:")
			conn.Close()
			break
		}
		
		if err!=nil||len<0{
			fmt.Println("readlength less than -1:")
			if err==syscall.EINTR{
				continue
			}
			break
		}
		buf=make([]byte,4096)
		
		
		
	}
	
}

func Help(){
    //version(stderr);

	fmt.Fprintf(os.Stderr,"\n")
	fmt.Fprintf(os.Stderr," -d                            - directs command to the only connected USB device\n")
	fmt.Fprintf(os.Stderr,"                                 returns an error if more than one USB device is present.\n")
	fmt.Fprintf(os.Stderr," -e                            - directs command to the only running emulator.\n")
	fmt.Fprintf(os.Stderr,"                                 returns an error if more than one emulator is running.\n")
	fmt.Fprintf(os.Stderr," -s <serial number>            - directs command to the USB device or emulator with\n")
	fmt.Fprintf(os.Stderr,"                                 the given serial number. Overrides ANDROID_SERIAL\n")
	fmt.Fprintf(os.Stderr,"                                 environment variable.\n")
	fmt.Fprintf(os.Stderr," -p <product name or path>     - simple product name like 'sooner', or\n")
	fmt.Fprintf(os.Stderr,"                                 a relative/absolute path to a product\n")
	fmt.Fprintf(os.Stderr,"                                 out directory like 'out/target/product/sooner'.\n")
	fmt.Fprintf(os.Stderr,"                                 If -p is not specified, the ANDROID_PRODUCT_OUT\n")
	fmt.Fprintf(os.Stderr,"                                 environment variable is used, which must\n")
	fmt.Fprintf(os.Stderr,"                                 be an absolute path.\n")
	fmt.Fprintf(os.Stderr," devices                       - list all connected devices\n")
	fmt.Fprintf(os.Stderr," connect <host>[:<port>]       - connect to a device via TCP/IP\n")
	fmt.Fprintf(os.Stderr,"                                 Port 5555 is used by default if no port number is specified.\n")
	fmt.Fprintf(os.Stderr," disconnect [<host>[:<port>]]  - disconnect from a TCP/IP device.\n")
	fmt.Fprintf(os.Stderr,"                                 Port 5555 is used by default if no port number is specified.\n")
	fmt.Fprintf(os.Stderr,"                                 Using this command with no additional arguments\n")
	fmt.Fprintf(os.Stderr,"                                 will disconnect from all connected TCP/IP devices.\n")
	fmt.Fprintf(os.Stderr,"\n")
	fmt.Fprintf(os.Stderr,"device commands:\n")
	fmt.Fprintf(os.Stderr,"  adb push <local> <remote>    - copy file/dir to device\n")
	fmt.Fprintf(os.Stderr,"  adb pull <remote> [<local>]  - copy file/dir from device\n")
	fmt.Fprintf(os.Stderr,"  adb sync [ <directory> ]     - copy host->device only if changed\n")
	fmt.Fprintf(os.Stderr,"                                 (-l means list but don't copy)\n")
	fmt.Fprintf(os.Stderr,"                                 (see 'adb help all')\n")
	fmt.Fprintf(os.Stderr,"  adb shell                    - run remote shell interactively\n")
	fmt.Fprintf(os.Stderr,"  adb shell <command>          - run remote shell command\n")
	fmt.Fprintf(os.Stderr,"  adb emu <command>            - run emulator console command\n")
	fmt.Fprintf(os.Stderr,"  adb logcat [ <filter-spec> ] - View device log\n")
	fmt.Fprintf(os.Stderr,"  adb forward <local> <remote> - forward socket connections\n")
	fmt.Fprintf(os.Stderr,"                                 forward specs are one of: \n")
	fmt.Fprintf(os.Stderr,"                                   tcp:<port>\n")
	fmt.Fprintf(os.Stderr,"                                   localabstract:<unix domain socket name>\n")
	fmt.Fprintf(os.Stderr,"                                   localreserved:<unix domain socket name>\n")
	fmt.Fprintf(os.Stderr,"                                   localfilesystem:<unix domain socket name>\n")
	fmt.Fprintf(os.Stderr,"                                   dev:<character device name>\n")
	fmt.Fprintf(os.Stderr,"                                   jdwp:<process pid> (remote only)\n")
	fmt.Fprintf(os.Stderr,"  adb jdwp                     - list PIDs of processes hosting a JDWP transport\n")
	fmt.Fprintf(os.Stderr,"  adb install [-l] [-r] [-s] <file> - push this package file to the device and install it\n")
	fmt.Fprintf(os.Stderr,"                                 ('-l' means forward-lock the app)\n")
	fmt.Fprintf(os.Stderr,"                                 ('-r' means reinstall the app, keeping its data)\n")
	fmt.Fprintf(os.Stderr,"                                 ('-s' means install on SD card instead of internal storage)\n")
	fmt.Fprintf(os.Stderr,"  adb uninstall [-k] <package> - remove this app package from the device\n")
	fmt.Fprintf(os.Stderr,"                                 ('-k' means keep the data and cache directories)\n")
	fmt.Fprintf(os.Stderr,"  adb bugreport                - return all information from the device\n")
	fmt.Fprintf(os.Stderr,"                                 that should be included in a bug report.\n")
	fmt.Fprintf(os.Stderr,"\n")
	fmt.Fprintf(os.Stderr,"  adb backup [-f <file>] [-apk|-noapk] [-shared|-noshared] [-all] [-system|-nosystem] [<packages...>]\n")
	fmt.Fprintf(os.Stderr,"                               - write an archive of the device's data to <file>.\n")
	fmt.Fprintf(os.Stderr,"                                 If no -f option is supplied then the data is written\n")
	fmt.Fprintf(os.Stderr,"                                 to \"backup.ab\" in the current directory.\n")
	fmt.Fprintf(os.Stderr,"                                 (-apk|-noapk enable/disable backup of the .apks themselves\n")
	fmt.Fprintf(os.Stderr,"                                    in the archive; the default is noapk.)\n")
	fmt.Fprintf(os.Stderr,"                                 (-shared|-noshared enable/disable backup of the device's\n")
	fmt.Fprintf(os.Stderr,"                                    shared storage / SD card contents; the default is noshared.)\n")
	fmt.Fprintf(os.Stderr,"                                 (-all means to back up all installed applications)\n")
	fmt.Fprintf(os.Stderr,"                                 (-system|-nosystem toggles whether -all automatically includes\n")
	fmt.Fprintf(os.Stderr,"                                    system applications; the default is to include system apps)\n")
	fmt.Fprintf(os.Stderr,"                                 (<packages...> is the list of applications to be backed up.  If\n")
	fmt.Fprintf(os.Stderr,"                                    the -all or -shared flags are passed, then the package\n")
	fmt.Fprintf(os.Stderr,"                                    list is optional.  Applications explicitly given on the\n")
	fmt.Fprintf(os.Stderr,"                                    command line will be included even if -nosystem would\n")
	fmt.Fprintf(os.Stderr,"                                    ordinarily cause them to be omitted.)\n")
	fmt.Fprintf(os.Stderr,"\n")
	fmt.Fprintf(os.Stderr,"  adb restore <file>           - restore device contents from the <file> backup archive\n")
	fmt.Fprintf(os.Stderr,"\n")
	fmt.Fprintf(os.Stderr,"  adb help                     - show this help message\n")
	fmt.Fprintf(os.Stderr,"  adb version                  - show version num\n")
	fmt.Fprintf(os.Stderr,"\n")
	fmt.Fprintf(os.Stderr,"scripting:\n")
	fmt.Fprintf(os.Stderr,"  adb wait-for-device          - block until device is online\n")
	fmt.Fprintf(os.Stderr,"  adb start-server             - ensure that there is a server running\n")
	fmt.Fprintf(os.Stderr,"  adb kill-server              - kill the server if it is running\n")
	fmt.Fprintf(os.Stderr,"  adb get-state                - prints: offline | bootloader | device\n")
	fmt.Fprintf(os.Stderr,"  adb get-serialno             - prints: <serial-number>\n")
	fmt.Fprintf(os.Stderr,"  adb status-window            - continuously print device status for a specified device\n")
	fmt.Fprintf(os.Stderr,"  adb remount                  - remounts the /system partition on the device read-write\n")
	fmt.Fprintf(os.Stderr,"  adb reboot [bootloader|recovery] - reboots the device, optionally into the bootloader or recovery program\n")
	fmt.Fprintf(os.Stderr,"  adb reboot-bootloader        - reboots the device into the bootloader\n")
	fmt.Fprintf(os.Stderr,"  adb root                     - restarts the adbd daemon with root permissions\n")
	fmt.Fprintf(os.Stderr,"  adb usb                      - restarts the adbd daemon listening on USB\n")
	fmt.Fprintf(os.Stderr,"  adb tcpip <port>             - restarts the adbd daemon listening on TCP on the specified port")
	fmt.Fprintf(os.Stderr,"\n")
	fmt.Fprintf(os.Stderr,"networking:\n")
	fmt.Fprintf(os.Stderr,"  adb ppp <tty> [parameters]   - Run PPP over USB.\n")
	fmt.Fprintf(os.Stderr," Note: you should not automatically start a PPP connection.\n")
	fmt.Fprintf(os.Stderr," <tty> refers to the tty for PPP stream. Eg. dev:/dev/omap_csmi_tty1\n")
	fmt.Fprintf(os.Stderr," [parameters] - Eg. defaultroute debug dump local notty usepeerdns\n")
	fmt.Fprintf(os.Stderr,"\n")
	fmt.Fprintf(os.Stderr,"adb sync notes: adb sync [ <directory> ]\n")
	fmt.Fprintf(os.Stderr,"  <localdir> can be interpreted in several ways:\n")
	fmt.Fprintf(os.Stderr,"\n")
	fmt.Fprintf(os.Stderr,"  - If <directory> is not specified, both /system and /data partitions will be updated.\n")
	fmt.Fprintf(os.Stderr,"\n")
	fmt.Fprintf(os.Stderr,"  - If it is \"system\" or \"data\", only the corresponding partition\n")
	fmt.Fprintf(os.Stderr,"    is updated.\n")
	fmt.Fprintf(os.Stderr,"\n")
	fmt.Fprintf(os.Stderr,"environmental variables:\n")
	fmt.Fprintf(os.Stderr,"  ADB_TRACE                    - Print debug information. A comma separated list of the following values\n")
	fmt.Fprintf(os.Stderr,"                                 1 or all, adb, sockets, packets, rwx, usb, sync, sysdeps, transport, jdwp\n")
	fmt.Fprintf(os.Stderr,"  ANDROID_SERIAL               - The serial number to connect to. -s takes priority over this if given.\n")
	fmt.Fprintf(os.Stderr,"  ANDROID_LOG_TAGS             - When used with the logcat option, only these debug tags are printed.\n")
}
