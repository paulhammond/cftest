# cftest

cftest is a utility to run tests against [Cloudfront Functions][].

Amazon helpfully provides a [Cloudfront TestFunction API][testfunction], which
allows you to get the output from running a Cloudfront Function with provided
input values. However that API (and the associated
`aws cloudfront test-function` command line tool) do not provide the ability to
easily test a function with multiple inputs and ensure all of them generate the
expected outputs. That's where cftest comes in.

[Cloudfront functions]: https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/cloudfront-functions.html
[testfunction]: https://docs.aws.amazon.com/cloudfront/latest/APIReference/API_TestFunction.html

## Installation

Prebuilt binaries can be downloaded from from
[Github Releases](https://github.com/paulhammond/cftest/releases)

## Usage

cftest works with a directory of JSON files. Each one is a separate test case
and specifies the input event and the expected output. Both are specified using
the [Cloudfront Function event structure][event]. For example the following
file specifies that a request to `/` should be redirected to `/foo`

```json
{
  "event": {
    "request": {
      "method": "GET",
      "uri": "/",
      "querystring": {},
      "headers": {
        "host": {
          "value": "example.com"
        }
      },
      "cookies": {}
    }
  },

  "output": {
    "response": {
      "headers": {
        "location": {
          "value": "https://example.com/foo"
        }
      },
      "statusDescription": "Found",
      "cookies": {},
      "statusCode": 302
    }
  }
}
```

Once you have a set of files and a cloudfront function, you can pass both to
cftest as arguments:

```
cftest myfunction:DEVELOPMENT tests/one.json tests/two.json
```

You can test both the `LIVE` or `DEVELOPMENT` stage of your function.

cftest uses the same authentication mechanisms as other AWS tools, including
instance roles, ~/.aws files and associated environment variables.

[event]: https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/functions-event-structure.html

## Error checking

Cloudfront functions can return two kinds of errors. The first is a HTTP error
with a code, such as "404 Not Found". The second is a thrown error. cftest can
test both.

To test a function returns a HTTP error, check the output matches the expected
HTTP response. For example:

```json
{
  "event": {…},

  "output": {
    "response": {
      "headers": {},
      "statusDescription": "Not Found",
      "cookies": {},
      "statusCode": 404
    }
  }
}
```

To test for a thrown error, specify an `error` instead of an `output`. For
example:

```json
{
  "event": {…},
  "error": "My thrown error"
}
```

## License

cftest is licensed under the [MIT license](LICENSE). For information on the
licenses of all included modules run `cftest --credits`.
