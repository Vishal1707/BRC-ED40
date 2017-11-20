# BRC-ED40
BRC-ED40 is linux websocket service for ED40 barcode scanner.  It consists of a single binary that can be installed as a service or daemon on linux. Log output goes to the syslog under linux.

On detection of barcode service sends 'Code' as json object via websocket.

## Requirements
The BRC-ED40 has been developed for linux. It has been developed and tested on GO1.8.

I assume you have a working Golang environment. As go get doesn't work with private repositories over https, we manually set up the project. Go to your Golang workspace and create the project path under the src folder:

```
cd src
mkdir -p brc-ed40
cd brc-ed40
```

After cloning the repositry, please switch to the main repositry folder and build the service,

```
go build
```

A binary named BRC-ED40 will be created in the same folder.

## Dependencies
### Libraries

All libraries/packages used by the BRC-ED40 service are managed using <a href="https://glide.sh/">glide</a>

After the initial clone all dependencies are resolved and can be found in the vendor folder. However, if you need to update dependencies, issue the following glide commands:
```
glide cc
glide install
glide update
```

## Usage

The service can be started on the console just by typing

```
BRC-ED40
```

It will then listen on localhost:8080 for websocket connections. Note that you need to provide a certificate and private key for TLS.

In order to start the service for testing purposes without TLS, use the --notls argument.

The host and port can be changed via --host and --port arguments, i.e.

```
BRC-ED40 --host=127.0.0.1 --port=6666 --notls
```

or with TLS:

```
BRC-ED40 --host=127.0.0.1 --port=6666 --cert=mycert.pem --key=myprivkey.pem
```
In order to install the service as a daemon use the --service argument along with the other preferred arguments, such as host and port:

```
./BRC-ED40 --host=127.0.0.1 --port=6666 --notls --service=install
```
In order to start the service and stop it once it is installed just use:

```
service BRC-ED40 start
```
and 
```
service BRC-ED40 stop
```

To uninstall the service just use:

```
BRC-ED40 --service=uninstall
```

Note: You need administrative rights to install/uninstall the service. 
