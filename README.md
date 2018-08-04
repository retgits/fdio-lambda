# fdio-lambda - Flogo Dot IO for AWS Lambda

A serverless tool designed to help create the `items.toml` from which the [showcase](https://tibcosoftware.github.io/flogo/showcases/) and the [flogo cli](https://github.com/TIBCOSoftware/flogo-cli) can get their search results.

## Layout
```bash
.
├── Makefile                    <-- Makefile to build and deploy
├── event.json                  <-- Sample event to test using SAM local
├── README.md                   <-- This file
├── src                         <-- Source code for a lambda function
│   ├── main.go                 <-- Lambda trigger code
│   ├── s3Util.go               <-- Utils to interact with Amazon S3
│   └── ssmUtil.go              <-- Utils to interact with Amazon SSM
└── template.yaml               <-- SAM Template
```

## Todo
- [ ] Create a more restrictive policy set for the S3 capabilities