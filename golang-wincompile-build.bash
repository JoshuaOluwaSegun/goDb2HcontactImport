#!/bin/bash
# Orignal https://gist.github.com/jmervine/7d3f455e923cf2ac3c9e
# usage: ./golang-crosscompile-build.bash

#Get current working directory
currentdir=`pwd`

#Clear Sceeen
 clear

# Get Version out of target then replace . with _
versiond=$(go run *.go -version)
version=${versiond//./_}

#Remove White Space
version=${version// /}
versiond=${versiond// /}

printf " ---- Building SQL Contact Import $versiond For Windows ---- \n"

package=goHornbillContactImport

printf "\n"
output386="contactImport_x86.exe"
outputx64="contactImport_x64.exe"

printf "Platform: Windows - 386\n"
destination="builds/windows/$output386"
printf "Go Build\n"
GOOS=windows GOARCH=386 go build  -o $destination

printf "\n"
printf "Platform: Windows - amd64 \n"
destination="builds/windows/$outputx64"
printf "Go Build\n"
GOOS=windows GOARCH=amd64 go build  -o $destination

printf "\n"
printf "Copy Source Files\n"
cp LICENSE.md "builds/windows/LICENSE.md"
cp README.md "builds/windows/README.md"
cp conf.json "builds/windows/conf.json"

printf "Build Zip \n"
cd "builds/windows/"
zip -r "${package}_windows_v${version}.zip" $output386 $outputx64 LICENSE.md README.md conf.json > /dev/null
cp "${package}_windows_v${version}.zip" "../../${package}_windows_v${version}.zip"

printf "\n"
printf "Clean Up \n"
cd $currentdir
\rm -rf "builds/"
printf "Build Complete \n"
printf "\n"