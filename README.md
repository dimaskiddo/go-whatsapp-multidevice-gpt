# Go WhatsApp Multi-Device Implementation as ChatGPT Bot

This repository contains example of implementation [go.mau.fi/whatsmeow](https://go.mau.fi/whatsmeow/). This example is using a codebase from [dimaskiddo/codebase-go-cli](https://github.com/dimaskiddo/codebase-go-cli).

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.
See deployment section for notes on how to deploy the project on a live system.

### Prerequisites

Prequisites packages:
* Go (Go Programming Language)
* GoReleaser (Go Automated Binaries Build)
* Make (Automated Execution using Makefile)

Optional packages:
* Docker (Application Containerization)

### Deployment

#### **Using Container**

1) Install Docker CE based on the [manual documentation](https://docs.docker.com/desktop/)

2) Run the following command on your Terminal or PowerShell
```sh
docker run -d \
  -e CHATGPT_API_KEY=<OPENAI_API_KEY> \
  --name go-whatsapp-multidevice-gpt \
  --rm dimaskiddo/go-whatsapp-multidevice-gpt:latest
```

3) Run the following command to Generate QR Code to be Scanned on your WhatsApp mobile application
```sh
docker exec -it \
  go-whatsapp-multidevice-gpt \
  go-whatsapp-multidevice-gpt login
```

#### **Using Pre-Build Binaries**

1) Download Pre-Build Binaries from the [release page](https://github.com/dimaskiddo/go-whatsapp-multidevice-gpt/releases)

2) Extract the zipped file

3) Copy the `.env.default` file as `.env` file

4) Run the pre-build binary
```sh
# MacOS / Linux
chmod 755 go-whatsapp-multidevice-gpt
./go-whatsapp-multidevice-gpt help
./go-whatsapp-multidevice-gpt login
./go-whatsapp-multidevice-gpt daemon

# Windows
# You can double click it or using PowerShell
.\go-whatsapp-multidevice-gpt.exe help
.\go-whatsapp-multidevice-gpt.exe login
.\go-whatsapp-multidevice-gpt.exe deamon
```

#### **Build From Source**

Below is the instructions to make this source code running:

1) Create a Go Workspace directory and export it as the extended GOPATH directory
```sh
cd <your_go_workspace_directory>
export GOPATH=$GOPATH:"`pwd`"
```

2) Under the Go Workspace directory create a source directory
```sh
mkdir -p src/github.com/dimaskiddo/go-whatsapp-multidevice-gpt
```

3) Move to the created directory and pull codebase
```sh
cd src/github.com/dimaskiddo/go-whatsapp-multidevice-gpt
git clone -b master https://github.com/dimaskiddo/go-whatsapp-multidevice-gpt.git .
```

4) Run following command to pull vendor packages
```sh
make vendor
```

5) Link or copy environment variables file
```sh
ln -sf .env.example .env
# - OR -
cp .env.example .env
```

6) Until this step you already can run this code by using this command
```sh
make run
```

7) *(Optional)* Use following command to build this code into binary spesific platform
```sh
make build
```

8) *(Optional)* To make mass binaries distribution you can use following command
```sh
make release
```

## Running The Tests

Currently the test is not ready yet :)

## Built With

* [Go](https://golang.org/) - Go Programming Languange
* [GoReleaser](https://github.com/goreleaser/goreleaser) - Go Automated Binaries Build
* [Make](https://www.gnu.org/software/make/) - GNU Make Automated Execution
* [Docker](https://www.docker.com/) - Application Containerization

## Authors

* **Dimas Restu Hidayanto** - *Initial Work* - [DimasKiddo](https://github.com/dimaskiddo)

See also the list of [contributors](https://github.com/dimaskiddo/go-whatsapp-multidevice-gpt/contributors) who participated in this project

## Annotation

You can seek more information for the make command parameters in the [Makefile](https://github.com/dimaskiddo/go-whatsapp-multidevice-gpt/-/raw/master/Makefile)

## License

Copyright (C) 2022 Dimas Restu Hidayanto

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
