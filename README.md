ksuidx
------------------------------------------------
`ksuidx` is an extension to the excellent `ksuid` library by Segment that allows
includes a resource identifier in the uid.

### Motivation

The primary driver is to allow ksuid values to be annotated with a namespace
to allow the type of record associated with the ksuid to be easily identified.

For example:

* the `org` might represent an organization
* the `usr` could represent a user
* ... and so on and so forth  

