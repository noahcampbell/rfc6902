golang JSON Patch (RFC 6902) Library 

### What is it?

RFC 6902 - JSON Patch - http://tools.ietf.org/html/rfc6902 defines how to properly parse a JSON Patch document and apply it to an existing JSON blob.  This library provides a way to parse those patch documents and transform a JSON document.

### The Latest Version

***The current version does not handle nested arrays very well.***

Check out https://github.com/noahcampbell/rfc6902/releases for all offical releases.

### Usage

    patch, err := ParsePatch(...)
    jsonDocTransformed, err := patch.Apply(jsonDoc)


### Documentation

The documentation is available via godoc.

### Installation

    go get github.com/noahcampbell/rfc6902

### Licensing

Apache License.  Please see the file called LICENSE for complete license.

### Contacts

Please submit any issues here: https://github.com/noahcampbell/rfc6902/issues

### Contributing

Pull requests will be accepted with accompanying unit tests.

