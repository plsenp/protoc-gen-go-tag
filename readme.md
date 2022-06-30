# protoc-gen-go-tag
rewrite *pb.go file insert tag

usage:
```shell
protoc-gen-go-tag example1.pb.go example2.pb.go
```
```protobuf
......
message Example {
    // @tag: uri; binding:required
    int64 id = 1;
    // @tag: binding:min=5,max=10
    string name =2;
}
```

```go
// xx.pb.go

type Example struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @tag: uri;xml;yml;json:override,id;form:uid
	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id"  uri:"id" xml:"id" yml:"id" form:"uid"`
	// @tag: uri: uname; xml;json:test,ttag
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty,test,ttag" uri:"uname"  xml:"name"`
	// @tag: binding:gt=0,lt=200
	Age int64 `protobuf:"varint,3,opt,name=age,proto3" json:"age,omitempty" binding:"gt=0,lt=200"`
}
```
