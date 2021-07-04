# protoc-gen-defaults

*This project is currently in **alpha**. The API should be considered unstable and likely to change*

**protoc-gen-defaults** is a protoc plugin generating the implementation of a `Defaulter` 
interface on messages:
```go
type Defaulter interface {
	Default()
}
```

## Installation

```bash
go get go.linka.cloud/protoc-gen-defaults
```

## Overview

**protoc-gen-defaults** make use of **Protobuf** options to define defaults field value.


Protobuf **Well-Known Types** are fully supported.

*TODO: Details*

## TODO
- [ ] docs
- [x] oneof support
- [ ] repeated support
- [ ] maps support
- [ ] bytes support
- [ ] add more generic methods to use as default value, e.g. *uuid*, *bsonid*... ?
