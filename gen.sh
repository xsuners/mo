#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

out=../generated/go

function gen(){
	#for file in $(ls $1)
    for file in `ls $1`
    do
        if [ -d $1"/"$file ]
        then
            if ls $1"/"$file"/"*.proto 1> /dev/null 2>&1;
            then
                protoc --proto_path=. \
                --go_out=$out \
                --go_opt=paths=source_relative \
                --go-mo_out=require_unimplemented_servers=false:$out \
                --go-mo_opt=paths=source_relative \
                $1"/"$file"/"*.proto
                echo "ok "$1"/"$file
            else
                echo "no "$1"/"$file
            fi
            gen $1"/"$file	#遍历子目录
        fi
    done
}

function clean_omitempty(){
	#for file in $(ls $1)
    for file in `ls $1`
    do
        if [ -d $1"/"$file ]
        then
            if ls $1"/"$file"/"*.go 1> /dev/null 2>&1;
            then
                # echo $1"/"$file"/"*.go
                sed -i -e "s/,omitempty//g" $1"/"$file"/"*.go
                echo "sed ok "$1"/"$file
            else
                echo "sed no "$1"/"$file
            fi
            clean_omitempty $1"/"$file	#遍历子目录
        fi
    done
}

cd $1

echo "gen folders:"

rm -rf $out/*
gen . $out

clean_omitempty $out
