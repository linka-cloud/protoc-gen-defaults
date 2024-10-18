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

## Usage

### Overview

**protoc-gen-defaults** make use of **Protobuf** options to define defaults field value.

### Generation

**protoc-gen-defaults** works the same way does the **protoc** plugins

Example:
```bash
protoc -I. -I defaults --go_out=paths=source_relative:. --defaults_out=paths=source_relative:. types.proto
```

### Disable generation or implementation

Implementation generation can be ignored with the `(defaults.ignored) = true` message option.

```proto
message NoDefaulterImplementation {
	option (defaults.ignored) = true;
	string string_field = 1 [(defaults.value).string = "string_field"];
}
```
> It may be useful if you intend to write your own `Defaulter` implementation.


An empty implementation can be generated with the `(defaults.disabled) = true` message option

```proto
message EmptyDefaulterImplementation {
    option (defaults.disabled) = true;
    string string_field = 1 [(defaults.value).string = "string_field"];
}
```

### Scalar and Well-Known Value

Each scalar or Well-Known type has its corresponding `(defaults.value).[scalar] = [value]` option, 
the `[value]` will be set if the scalar field has the **zero value**, e.g `0` for numbers, 
`""` for strings, `false` for bools

- **float / google.protobuf.FloatValue**:
    ```proto
    float float = 1 [(defaults.value).float = 0.42];
    ```
- **double / google.protobuf.DoubleValue**: 
    ```proto 
    double double = 2 [(defaults.value).double = 0.42];
    ****google.protobuf.DoubleValue double_value = 20 [(defaults.value).double = 0.42];
    ```
- **int32 / google.protobuf.Int32Value**: 
    ```proto  
    int32 int32 = 3 [(defaults.value).int32 = 42];
    google.protobuf.Int32Value int32_value = 24 [(defaults.value).int32 = 42];
    ```
- **int64 / google.protobuf.Int64Value**: 
    ```proto  
    int64 int64 = 4 [(defaults.value).int64 = 42];
    google.protobuf.Int64Value int64_value = 22 [(defaults.value).int64 = 42];
    ```
- **uint32 / google.protobuf.UInt32Value**: 
    ```proto  
    uint32 uint32 = 5 [(defaults.value).uint32 = 42];
    google.protobuf.UInt32Value uint32_value = 25 [(defaults.value).uint32 = 42];
    ```
- **uint64 / google.protobuf.UInt64Value**: 
    ```proto  
    uint64 uint64 = 6 [(defaults.value).uint64 = 42];
    google.protobuf.UInt64Value uint64_value = 23 [(defaults.value).uint64 = 42];
    ```
- **sint32**: 
    ```proto  
    sint32 sint32 = 7 [(defaults.value).sint32 = 42];
    ```
- **sint64**: 
    ```proto  
    sint64 sint64 = 8 [(defaults.value).sint64 = 42];
    ```
- **fixed32**: 
    ```proto  
    fixed32 fixed32 = 9 [(defaults.value).fixed32 = 42];
    ```
- **fixed64**: 
    ```proto  
    fixed64 fixed64 = 10 [(defaults.value).fixed64 = 42];
    ```
- **sfixed32**: 
    ```proto  
    sfixed32 sfixed32 = 11 [(defaults.value).sfixed32 = 42];
    ```
- **sfixed64**: 
    ```proto  
    sfixed64 sfixed64 = 12 [(defaults.value).sfixed64 = 42];
    ```
- **bool / google.protobuf.BoolValue**: 
    ```proto  
    bool bool = 13 [(defaults.value).bool = true];
    google.protobuf.BoolValue bool_value = 26 [(defaults.value).bool = false];
    ```
- **string / google.protobuf.StringValue**: 
    ```proto  
    string string = 14 [(defaults.value).string = "42"];
    google.protobuf.StringValue string_value = 27 [(defaults.value).string = "42"];
    ```
- **bytes / google.protobuf.BytesValue**: 
    ```proto  
    bytes bytes = 15 [(defaults.value).bytes = "42"];
    google.protobuf.BytesValue bytes_value = 28 [(defaults.value).bytes = "42"];
    ```
    
### Messages 

Message default behaviour is defined with the `(defaults.value).message = {initialize: bool, defaults: bool}` field option.

- `initialize`: if set to `true` the field will be initialized with an empty `stuct reference` from the appropriate type
- `defaults`: tells that the `Default` method should be called if the type implements the `Defaulter` interface

```proto
Message message = 17 [(defaults.value).message = {initialize: true, defaults: true}];
```

### Well-Known Messages

**google.protobuf.Duration** 
  
The default value is parsed according to the [Prometheus time durations format](https://prometheus.io/docs/prometheus/latest/querying/basics/#time-durations)
```proto
google.protobuf.Duration duration = 18 [(defaults.value).duration = "2d"];
```

**google.protobuf.Timestamp**

The default value is parsed using the following RFCs as defined in the `time` package:
  - time.RFC822
  - time.RFC822Z
  - time.RFC850
  - time.RFC1123
  - time.RFC1123Z
  - time.RFC3339

Timestamp also support a convenient value `now` which will set the value from `time.Now()` at `Default()` method call time.

```proto
google.protobuf.Timestamp timestamp = 19 [(defaults.value).timestamp = "now"];
google.protobuf.Timestamp time_value_field_with_default = 18 [(defaults.value).timestamp = "1952-03-11T00:00:00Z"];
```

### Enums

The enum index value if currently to zero.

```proto
enum Enum {
    NONE = 0;
    ONE = 1;
    TWO = 2;
}
Enum enum = 16 [(defaults.value).enum = 1];
```

### oneof

If the `defaults.oneof` option is set, the `oneof` will be initialized with a struct of the *oneof type wrapper*, 
the regular field type `default` option will be applied.

```proto
oneof one_of {
    option (defaults.oneof) = "two";
    OneOfOne one = 29 [(defaults.value).message = {defaults: true, initialize: true}];
    OneOfTwo two = 30 [(defaults.value).message = {defaults: true, initialize: true}];
    OneOfThree three = 31 [(defaults.value).message = {defaults: true, initialize: true}];
    Enum four = 32 [(defaults.value).enum = 1];
}
```


### Repeated and Maps

`repeated` and `maps` are not supported.


## All types example
```proto
syntax = "proto3";

package tests;

option go_package = "go.linka.cloud/protoc-gen-defaults/tests/pb;pb";

import "defaults/defaults.proto";

import "google/protobuf/wrappers.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/duration.proto";

message Types {
    // Scalar Field Types
    float float = 1 [(defaults.value).float = 0.42];
    double double = 2 [(defaults.value).double = 0.42];
    int32 int32 = 3 [(defaults.value).int32 = 42];
    int64 int64 = 4 [(defaults.value).int64 = 42];
    uint32 uint32 = 5 [(defaults.value).uint32 = 42];
    uint64 uint64 = 6 [(defaults.value).uint64 = 42];
    sint32 sint32 = 7 [(defaults.value).sint32 = 42];
    sint64 sint64 = 8 [(defaults.value).sint64 = 42];
    fixed32 fixed32 = 9 [(defaults.value).fixed32 = 42];
    fixed64 fixed64 = 10 [(defaults.value).fixed64 = 42];
    sfixed32 sfixed32 = 11 [(defaults.value).sfixed32 = 42];
    sfixed64 sfixed64 = 12 [(defaults.value).sfixed64 = 42];
    bool bool = 13 [(defaults.value).bool = true];
    string string = 14 [(defaults.value).string = "42"];
    bytes bytes = 15 [(defaults.value).bytes = "42"];

    // Complex Field Types
    enum Enum {
        NONE = 0;
        ONE = 1;
        TWO = 2;
    }
    Enum enum = 16 [(defaults.value).enum = 1];
    Message message = 17 [(defaults.value).message = {initialize: true, defaults: false}];
    oneof one_of {
        option (defaults.oneof) = "two";
        OneOfOne one = 29 [(defaults.value).message = {defaults: true, initialize: true}];
        OneOfTwo two = 30 [(defaults.value).message = {defaults: true, initialize: true}];
        OneOfThree three = 31 [(defaults.value).message = {defaults: true, initialize: true}];
        Enum four = 32 [(defaults.value).enum = 1];
    }
    
    // WellKnow types
    google.protobuf.Duration duration = 18 [(defaults.value).duration = "2d"];
    google.protobuf.Timestamp timestamp = 19 [(defaults.value).timestamp = "now"];
    google.protobuf.DoubleValue double_value = 20 [(defaults.value).double = 0.42];
    google.protobuf.FloatValue float_value = 21 [(defaults.value).float = 0.42];
    google.protobuf.Int64Value int64_value = 22 [(defaults.value).int64 = 42];
    google.protobuf.UInt64Value uint64_value = 23 [(defaults.value).uint64 = 42];
    google.protobuf.Int32Value int32_value = 24 [(defaults.value).int32 = 42];
    google.protobuf.UInt32Value uint32_value = 25 [(defaults.value).uint32 = 42];
    google.protobuf.BoolValue bool_value = 26 [(defaults.value).bool = false];
    google.protobuf.StringValue string_value = 27 [(defaults.value).string = "42"];
    google.protobuf.BytesValue bytes_value = 28 [(defaults.value).bytes = "42"];
}

message Message {
    string field = 1 [(defaults.value).string = "lonely field"];
}

message OneOfOne {
    option (defaults.ignored) = true;
    string string_field = 1 [(defaults.value).string = "string_field"];
}

message OneOfTwo {
    string string_field = 1 [(defaults.value).string = "string_field"];
}

message OneOfThree {
    option (defaults.disabled) = true;
    string string_field = 1 [(defaults.value).string = "string_field"];
}

```

### Using reflection / without code generation

Setting protobuf message defaults is also supported using reflection:

```go
package main

import (
	pb "..."
	"go.linka.cloud/protoc-gen-defaults/defaults"
)

func main() {
	var msg pb.MyMessage
	defaults.Apply(&msg)
}

```

## TODO
- [x] docs
- [x] oneof support
- [x] set default values by using [Protobuf reflection](https://pkg.go.dev/google.golang.org/protobuf@v1.27.1/reflect/protoreflect)
- [ ] add more generic methods to use as default value, e.g. *uuid*, *xid*, *bsonid*... ?
- [ ] repeated support ?
- [ ] maps support ?
- [x] bytes support
