package jdk
import(
	"os/exec"
	"strings"
	"regexp"
	"fmt"
	"io/ioutil"
	"../file"
)

//add a new method to our new regular expression type
func FindStringSubmatchMap(reg string,s string) map[string]string{
	captures:=make(map[string]string)
	r := regexp.MustCompile(reg)
	match:=r.FindStringSubmatch(s)
	if match==nil{
		return captures
	}
	for i,name:=range r.SubexpNames(){
		//Ignore the whole regexp match and unnamed groups
		if i==0||name==""{
			continue
		}
		captures[name]=match[i]
	}
	return captures
}

/**
 * Returns version, architecture
 */
func GetCurrentVersion() (string, string) {

	cmd := exec.Command("java","-version")
	str, err := cmd.Output()
	if err == nil {
		fmt.Println(string(str))
		mmap := FindStringSubmatchMap(`java version "(?P<version>.*)"`,string(str))

		v := mmap["version"]
		isX64 := strings.IndexAny(string(str),"64-Bit") > -1
		if isX64 {
				return v, "x64"
		}else {
			return v, "x86"
		}
	}
	return "Unknown",""
}

func GetInstalled(root string) []string {
	list := make([]string,0)
	files, _ := ioutil.ReadDir(root)
	for i := len(files) - 1; i >= 0; i-- {
		if files[i].IsDir() {
			isnode, _ := regexp.MatchString("^v",files[i].Name())
			if isnode {
				list = append(list,files[i].Name())
			}
		}
	}
	return list
}

func IsVersionInstalled(root string, version string, cpu string) bool {
	fileName := file.GenJDKFileName(version,cpu)
	isInstalled := file.Exists(root+"\\v"+fileName+"\\bin\\javac.exe")
	fmt.Print(root+"\\v"+fileName+"\\bin\\javac.exe  ")
	fmt.Println(isInstalled)
	return isInstalled
}