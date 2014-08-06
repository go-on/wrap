# v2.0 

## Breaking changes

Some adapters had some confusing namings around combinations of the words 
Handle(r) Serve(r) and Wrapper. 

The following adapters were renamed, since there were confusing:

ServeWrapper => NextHandler
ServeHandlerFunc => NextHandlerFunc

The ServeHandler interface and the ServeHandleFunc were removed.

These changes affect the go-on/wrap-contrib and go-on/wrap-contrib-testing repositories which have been changed to be again compatible with wrap.

Apart from that not many users should be affected.

## New features

- Contexter interface for unified context sharing between middleware
- support for debugging
- improved documentation with examples
- response writer wrappers to help the creation of middlewares
- helpers to extract the original ResponseWriter out of a Contexter


# v1.0

## Features

- Wrapper interface for middleware
- New function to create a stack of Wrappers
- adapters
- benchmark