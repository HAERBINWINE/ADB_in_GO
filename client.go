package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	//"flags"
	"syscall"
	"bytes"
	"time"
)

func main() {
	interactive_shell()
}
func custom(){
	//打开连接:
	conn, err := net.Dial("tcp", "localhost:5037")
	defer conn.Close()
	if err != nil {
		//由于目标计算机积极拒绝而无法创建连接
		fmt.Println("Error dialing", err.Error())
		return // 终止程序
	}

	inputReader := bufio.NewReader(os.Stdin)

	// 给服务器发送信息直到程序退出：
	go read(conn)
	for conn!=nil {
		fmt.Println("What to send to the server? Type Q to quit.")
		input, _ := inputReader.ReadString('\n')
		trimmedInput := strings.Trim(input, "\r\n")
		// fmt.Printf("input:--s%--", input)
		trimmedInput=fmt.Sprintf("%04x%s",len(trimmedInput),trimmedInput)
		fmt.Printf("trimmedInput:--%s--", trimmedInput)
		
		if trimmedInput == "Q" {
			return
		}
		len, err := conn.Write([]byte(trimmedInput))
		fmt.Println(len)
		if len<0{
			if err!=nil{
				fmt.Println("Write shell cmd failed,err:",err)
			}
			conn.Close()
		}
	}
}

func interactive_shell()int{
	
	conn:=adb_connect([]byte("shell:"))
	if conn==nil{
		fmt.Println("failed to make shell")
		return 1
	}
	 read(conn)
	//go writeToShell(conn)
	
	return 0
}
func switch_socket_transport(conn net.Conn,sn string) int{
	fmt.Println("switch_socket_transport")
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
func adb_connect(c []byte) net.Conn{
	conn:=_adb_connect([]byte("host:version"))
	if conn==nil{
		fmt.Println("failed to start daemon")
	}else{
		buf:=make([]byte,16)
		conn.Read(buf)
		fmt.Println("adb version:",string(buf[8:]))
		conn.Close()
	}
	conn=_adb_connect(c)
	
	if conn==nil{
		fmt.Println("failed to start service")
		return nil
	}
	return conn
}
func _adb_connect(c []byte) net.Conn{
	
	len:=bytes.Count(c,nil)-1
	if len<1 || len>1024  {
		fmt.Println("cmd too long")
		return nil
	}
	trimmedInput:=fmt.Sprintf("%04x%s",len,c)
	fmt.Println("trimmedInput: ", trimmedInput,string(c[:3]))
	conn:=create_socket()
	if conn==nil{
		return nil
	}
	if !bytes.Equal(c[:4],[]byte("host")) && switch_socket_transport(conn,"fa05d040")!=0 {
		
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
	for {
		fmt.Printf("# ")
		input, _ := inputReader.ReadString('\n')
		trimmedInput := strings.Trim(input, "\r\n")
		// fmt.Printf("input:--s%--", input)
		//trimmedInput=fmt.Sprintf("%04x%s",len(trimmedInput),trimmedInput)
		//fmt.Printf("trimmedInput:--%s--", trimmedInput)
		
		if trimmedInput == "Q" {
			return
		}
		time.Sleep(1)
		_, err := conn.Write([]byte(trimmedInput+"\n"))
		
		if err !=nil{
			fmt.Printf("write shell command failed\n")
			conn.Close()
		}
	}
}
func read(conn net.Conn) {
	buf:=make([]byte,4096)
	conn.SetReadDeadline(time.Now().Add(3000000000))
	for{
	for conn!=nil{
		conn.SetReadDeadline(time.Now().Add(1000000000))
		len,err:=conn.Read(buf)
		//fmt.Println("line %s",buf,len)
		time.Sleep(1)
		fmt.Printf(string(buf[:len]))
		if len==0{
			fmt.Println("no content")
			break
		}
		
		if len<0{
			if err==syscall.EINTR{
				continue
			}
			break
		}
		buf=make([]byte,4096)
		
		
		
		
	}
	fmt.Println(">>")
	_,err:=conn.Write([]byte("ls\n"))
	if err!=nil{
		fmt.Println(err)
	}
}
}