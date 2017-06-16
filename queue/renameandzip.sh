#!/bin/sh

for i in bin/*
do
  echo $i
  mv $i astqueue
  tar -cf $i.tar astqueue
  bzip2 -9 $i.tar
done

