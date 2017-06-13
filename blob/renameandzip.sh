#!/bin/sh

for i in bin/*
do
  echo $i
  mv $i astblob
  tar -cf $i.tar astblob
  bzip2 -9 $i.tar
done

