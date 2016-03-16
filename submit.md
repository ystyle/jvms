## Submit Other JDK Version URL

the new Version url will add this file [jdkversions.json](jdkversions.json)

### The filename structure
The filename include `jdk`+`$VERSION$`+`x86|x64`+`.zip`

All characters are lower case

e.g.  `jdk1.7.0.67_x64.zip`

### zip file structure
e.g:
```
$ tree -L 1 v1.7.0_67_x64/
v1.7.0_67_x64/
|-- COPYRIGHT
|-- LICENSE
|-- README.html
|-- THIRDPARTYLICENSEREADME-JAVAFX.txt
|-- THIRDPARTYLICENSEREADME.txt
|-- X64.txt
|-- bin
|-- db
|-- include
|-- jre
|-- lib
|-- release
`-- src.zip

5 directories, 8 files
```
