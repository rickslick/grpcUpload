# grpcUpload - Multi File Uploader

grpcUpload is CLI tool that uploads files in parallel using grpc.
Featues :
*  concurrent multi file upload using  grpc with concept of chunking
* supports tls (both client and sever )
* Displays progress for each file

![rkUploader](https://raw.githubusercontent.com/rickslick/grpcUpload/master/recording.gif)
## Usage

start the server :

```
$./grpcUploadServer serve --a localhost:9191
```

Upload all files in the specified directory to the server :

```
$ ./UploadClient upload  -a localhost:9191 -d <folder>
```

## License

MIT

## Author

Rohan Koshy (a.k.a. rickslick)
Based on multidownloader by mattn :  github.com/mattn/ft
