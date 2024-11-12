cpiostrip
=========

Sets modification time for all files and directories in cpio archive (`newc` only) to 1970-01-01 00:00:00

Can compare archives before and after stripping.

```
Usage: ./cpiostrip-linux-amd64 -in [FILE]
   or: ./cpiostrip-linux-amd64 -in [FILE] -out [FILE]
   or: ./cpiostrip-linux-amd64 -f1 [FILE] -f2 [FILE]

Reset modification time of cpio file and directory entries to Thu Jan  1 12:00:00 AM GMT 1970
Compare entries in archives and print fields that differ with -f1/-f2

  -f1 string
    	compare file 1
  -f2 string
    	compare file 2
  -in string
    	file to strip
  -out string
    	output file
```

### Build

```shell
make
```

### Example:

```shell
$ mkdir test
$ echo 123 > test/file
$ ls -la test
total 12
drwxr-xr-x 2 user user 4096 Apr 22 00:54 .
drwxr-xr-x 4 user user 4096 Apr 22 00:54 ..
-rw-r--r-- 1 user user    4 Apr 22 00:54 file

$ cd test && find . | cpio --create --format='newc' -O ../out.cpio && cd ..
1 block

$ cpio -ivt < out.cpio
drwxr-xr-x   2 user  user         0 Apr 22 00:54 .
-rw-r--r--   1 user  user         4 Apr 22 00:54 file
1 block

$ ./cpiostrip-linux-amd64 -in out.cpio 
2024/11/12 12:24:02 INFO stripping inplace filename=out.cpio
2024/11/12 12:24:02 INFO stripped filename=.
2024/11/12 12:24:02 INFO stripped filename=file

$ cpio -ivt < out.cpio
drwxr-xr-x   2 user  user         0 Jan  1  1970 .
-rw-r--r--   1 user  user         4 Jan  1  1970 file
1 block
```

Diff (modification time differs):

```shell
$ ./cpiostrip-linux-amd64 -f1 out.cpio -f2 ff.cpio 
file: . update ModTime:{0 63866993241 6704160} -> {0 62135596800 6704160}       
file: file      update ModTime:{0 63866993241 6704160} -> {0 62135596800 6704160}
```

### See also:

https://salsa.debian.org/reproducible-builds/strip-nondeterminism