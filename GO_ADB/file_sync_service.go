package GO_ADB

import(
	"net"
	"fmt"
	"unsafe"
	"reflect"
	"os"
	"bytes"
	"strings"
)

const S_IFSOCK_= 0140000
const S_IFLINK_= 0120000
const S_IFREG_ = 0100000
const S_IFBLK_ = 0060000
const S_IFDIR_ = 0040000
const S_IFCHR_ = 0020000
const S_IFIFO_ = 0010000

const SYNC_DATA_MAX=1024*64 //1round max 64KB

func ISREG(mode uint32) bool{
	return mode&S_IFREG_!=0
}
func ISLINK(mode uint32) bool{
	return mode&S_IFLINK_!=0
}
func ISCHR(mode uint32) bool{
	return mode&S_IFCHR_!=0
}
func ISBLK(mode uint32) bool{
	return mode&S_IFBLK_!=0
}
func ISDIR(mode uint32) bool{
	return mode&S_IFDIR_!=0
}

// big endian mode
func  __swap_uint32( x uint32)  uint32{
    return (((x) & 0xFF000000) >> 24)| (((x) & 0x00FF0000) >> 8)| (((x) & 0x0000FF00) << 8)| (((x) & 0x000000FF) << 24)
}
// little endian mode
func MKID(a,b,c,d byte) uint32{
	fmt.Printf("%x %x %x %x \n",d,c,b,a)
	return uint32(uint32(a)|(uint32(b)<<8)|(uint32(c)<<16)|(uint32(d)<<24))
}
func htoll(x uint32) uint32{
	return __swap_uint32(x)
}
func ltohl(x uint32) uint32{
	return __swap_uint32(x)
}
var ID_STAT=MKID('S','T','A','T')
var ID_LIST=MKID('L','I','S','T')
var ID_ULNK=MKID('U','L','N','K')
var ID_SEND=MKID('S','E','N','D')
var ID_RECV=MKID('R','E','C','V')
var ID_DENT=MKID('D','E','N','T')
var ID_DONE=MKID('D','O','N','E')
var ID_DATA=MKID('D','A','T','A')
var ID_OKAY=MKID('O','K','A','Y')
var ID_FAIL=MKID('F','A','I','L')
var ID_QUIT=MKID('Q','U','I','T')

/* Syncmsg, In C version, Syncmsg is a union
while in Go, to implement a Union will cost more,
*/

type Req struct {
	id uint32
	namelen uint32
} 

type Stat struct {
	id uint32
	mode uint32
	size uint32
	time uint32
} 
type Dent struct {
	id uint32
	mode uint32
	size uint32
	time uint32
	namelen uint32
} 
type Data struct {
	id uint32
	size uint32
} 
type Status struct {
	id uint32
	msglen uint32
}     
type  Syncmsg struct{
	id uint32
	req Req
	stat Stat
	dent Dent
	data Data
	status Status
} 


 type Syncsendbuf struct {
    id uint32
    size uint32;
    data [SYNC_DATA_MAX]byte;
}

var send_buffer Syncsendbuf 
func reqTobytes( req *Req) []byte{
	var temp reflect.SliceHeader
	voidptr:=uintptr(unsafe.Pointer(req))
	len:=int(unsafe.Sizeof(Req{}))
	temp.Data=voidptr
	temp.Len=len
	temp.Cap=len
	fmt.Printf("%d\n",len)
	return *(*[]byte)(unsafe.Pointer(&temp))
}

func bytesTostat(b *[]byte) Stat{
	return *(*Stat)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(b)).Data))
}

func bytesTodata(b *[]byte) Data{
	return *(*Data)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(b)).Data))
}

func DirStops(dir string) string{
	list1:=strings.Split(dir,"/")
	list2:=strings.Split(dir,"\\")
	fmt.Println(list1,list2)
	if len(list1)==0{
		return list2[len(list2)-1]
	}else{
		return list1[len(list1)-1]
	}
}
func sync_readmode(conn net.Conn,path string, mode *uint32) int{
	var msg  Syncmsg
	
	len:=len(path)
	msg.req.id=ID_STAT


	msg.req.namelen = uint32(len)
	
	buf:=reqTobytes(&msg.req)
	fmt.Printf("%b\n",buf)
	buffer_:=bytes.NewBuffer(buf)

	buffer_.Write([]byte(path))
	buf=buffer_.Bytes()
	fmt.Printf(">>>>>>%s\n",buf)
	_,err:=conn.Write(buf)
	if err!=nil{
		fmt.Fprintf(os.Stderr,"sync_readmode: write msg to server failed, err: %s\n",err)
		return -1
	}
	buf1:=make([]byte,512)
	len,err=conn.Read(buf1)
	fmt.Printf("STAT_PATH: %d:%s",len,buf1)
	if err!=nil{
		fmt.Fprintf(os.Stderr,"sync_readmode: read from server failed, err: %s\n",err)
		return -1
	}
	
	size:=int(unsafe.Sizeof(Stat{}))
	buf1=make([]byte,size)
	len,err=conn.Read(buf1)
	fmt.Printf("%d:%s",len,buf1)
	fmt.Println(buf1)
	if err!=nil{
		fmt.Fprintf(os.Stderr,"sync_readmode: read from server failed, err: %s\n",err)
		return -1
	}
	stat:=bytesTostat(&buf1)	
	fmt.Printf("%x %x %x",stat.id,ID_STAT,stat.mode)
	if stat.id!=ID_STAT{
		return -1
	}
	*mode=stat.mode
	return 0
}

func Sync_recv(conn net.Conn, src,dst string) int{
	var msg  Syncmsg
	len:=len(src)
	if len>1024{
		return -1
	}
	msg.req.id = ID_RECV
	msg.req.namelen = uint32(len)
	buf:=reqTobytes(&msg.req)
	fmt.Printf("%b\n",buf)
	buffer_:=bytes.NewBuffer(buf)

	buffer_.Write([]byte(src))
	buf=buffer_.Bytes()
	fmt.Printf(">>>>>>%s\n",buf)
	_,err:=conn.Write(buf)
	if err!=nil{
		fmt.Fprintf(os.Stderr,"Sync_recv: write msg to server failed, err: %s\n",err)
		return -1
	}

	size:=int(unsafe.Sizeof(Data{}))
	buf1:=make([]byte,size)
	len,err=conn.Read(buf1)
	fmt.Printf("%d:%s",len,buf1)
	fmt.Println(buf1)
	if err!=nil{
		fmt.Fprintf(os.Stderr,"Sync_recv: read from server failed, err: %s\n",err)
		return -1
	}
	data:=bytesTodata(&buf1)
	fmt.Printf("%d %d ",data.size,data.id)
	var file *os.File
	total_bytes:=0
	if data.id==ID_DATA || data.id == ID_DONE{
		os.Remove(dst)
		if strings.Contains(dst,"\\"){
			dstfolder:=strings.TrimRight(dst,"\\")
			os.MkdirAll(dstfolder,os.ModePerm)
		}else if strings.Contains(dst,"/"){
			dstfolder:=strings.TrimRight(dst,"/")
			os.MkdirAll(dstfolder,os.ModePerm)
		}

		file,err=os.Create(dst)
		if err!=nil{
			fmt.Fprintf(os.Stderr,"Sync_recv: create file %s failed, err: %s\n",dst,err)
		return -1
		}

	}else{
		goto remote_error
	}
	
	for{

		len=int(msg.data.size)
		if data.id == ID_DONE{
			break
		}
		if data.id != ID_DATA{
			fmt.Fprintf(os.Stderr,"Sync_recv: goto err handler, %x:%x",data.id,ID_DATA)
			goto remote_error
		}
		if(len>SYNC_DATA_MAX){
			fmt.Fprintf(os.Stderr,"Sync_recv: data over run")
			file.Close()
			return -1
		}

		buf_data:=make([]byte,len)
		len,err=conn.Read(buf_data)
		fmt.Printf("%d:%s",len,buf_data)
		fmt.Println(buf_data)
		if err!=nil{
			fmt.Fprintf(os.Stderr,"Sync_recv: pulling data failed, err: %s\n",err)
			return -1
		}

		len,err=file.Write(buf_data)
		if err!=nil{
			fmt.Fprintf(os.Stderr,"Sync_recv: Write to file  failed, err: %s\n",err)
			return -1
		}
		total_bytes+=len

		size=int(unsafe.Sizeof(Data{}))
		buf1=make([]byte,size)
		len,err=conn.Read(buf1)
		fmt.Printf("%d:%s",len,buf1)
		fmt.Println(buf1)
		if err!=nil{
			fmt.Fprintf(os.Stderr,"Sync_recv: read from server failed, err: %s\n",err)
			return -1
		}
		data=bytesTodata(&buf1)
		fmt.Printf("%d %d ",data.size,data.id)
	
	}
	fmt.Printf("file: %s, total : %d",file,total_bytes)
	file.Sync()
	file.Close()
	remote_error:
	return 0
}
func main(){
	//sync_readmode(nil,".",1)
}
