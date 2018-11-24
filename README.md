# grpcUpload - Multi File Uploader

grpcUpload is CLI tool that uploads files concurrently using grpc.
Featues :
*  concurrent multi file upload using  grpc with concept of chunking
* supports tls (both client and sever )
* Displays progress for each file

TODO
* retry handling on client per chunk / file
* Adding authorization grpc interceptor

![rkUploader](https://raw.githubusercontent.com/rickslick/grpcUpload/master/recording.gif)
## Usage

Server : start the server( default destination of files is /tmp)  :

```
$./grpcUploadServer serve --a <ip:port> -d <destination folder>
Eg ./UploadClient serve -a localhost:9191 -d /home/
```

Client : Upload all files in the specified directory to the server  :

```
$ ./UploadClient upload  -a <ip:port> -d <folder containing files to upload>   
Eg  ./UploadClient upload -a localhost:9191 -d /home/
```

## License

MIT

## Author

* Rohan Koshy (a.k.a. rickslick)
* Ascesh Dandey (a.k.a ascesh)
* Based on multidownloader by mattn :  github.com/mattn/ft
