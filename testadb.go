package main

import(
	"./GO_ADB"
	"os"
	"fmt"
)


func main(){
	var serialNumber string
	fmt.Println(os.Args)
	if len(os.Args)<2{
		GO_ADB.Help()
		return
	}
	switch os.Args[1]{
		case "-s":
			if len(os.Args)>=3{
				serialNumber=os.Args[2]
			}else{
				GO_ADB.Help()	
				return
			}
		case "pull":
			adb:=GO_ADB.NewAdb("")
			adb.Pull("","")
			return
			
	}
	fmt.Println(serialNumber)
	switch os.Args[3]{
		case "shell":
			adb:=GO_ADB.NewAdb(serialNumber)
			adb.Interactive_shell()
			return
		case "pull":
			adb:=GO_ADB.NewAdb(serialNumber)
			adb.Pull("/data/test/text.txt",".")
			return
	}
	
}