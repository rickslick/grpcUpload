# grpcUpload - Multi File Uploader

grpcUpload is CLI tool that uploads files in parallel using grpc.
Featues :
*  multi file upload using  grpc with concept of chunking
* supports tls (both client and sever )
* Displays progress for each file

## Usage

start the server :

```
$./grpcUploadServer serve --a localhost:9191
```

Upload all files in the specified directory to the server :

```
$ ./UploadClient upload  -a localhost:9191 -d <folder>
```

## Installation

```
$ go get github.com/rickslick/grpcUpload
```

## License

MIT

## Author

Rohan Koshy (a.k.a. rickslick)
Based on multidownloader by mattn :  github.com/mattn/ft
