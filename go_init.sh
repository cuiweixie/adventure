#!/bin/bash
set -e

#go version conf
GO_VERSION=(
"go1.16"
"go1.16.1"
"go1.16.2"
"go1.16.3"
"go1.16.4"
"go1.16.5"
"go1.16.6"
"go1.16.7"
"go1.16.8"
"go1.16.9"
"go1.16.10"
"go1.16.11"
"go1.16.12"
"go1.7"
"go1.7.1"
"go1.7.3"
"go1.7.4"
"go1.7.5"
"go1.7.6"
)

print_version(){
    length=`expr ${#GO_VERSION[*]} - 1`
    for ((i =${length}; i >= 0; i--))
      do
        echo ${i}"---->"${GO_VERSION[i]}
      done
    read -p "please enter num,choose go version:" num

  #Check if input is a number
  if echo $num | grep -q '[^0-9]';
    then
      echo "this is not a num,please enter num"

  elif [ $num -gt $length ];
    then
      echo "this number does not exist,please enter again"
  else
      # shellcheck disable=SC2104
      break
    fi
}


logo=1
while [ ${logo} ]
do
  print_version
done

#download and install
wget https://studygolang.com/dl/golang/${GO_VERSION[num]}.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf ${GO_VERSION[num]}.linux-amd64.tar.gz
sed -i "1iexport PATH=\$PATH:/usr/local/go/bin:/root/go/bin" ~/.bashrc
source ~/.bashrc
rm -f ${GO_VERSION[num]}.linux-amd64.tar.gz

apt-get update
apt install make gcc
make