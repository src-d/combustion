# combustion [![Build Status](https://travis-ci.org/src-d/combustion.svg?branch=master)](https://travis-ci.org/src-d/combustion)

Documentation
-------------

Combustion files are a extension to [container linux config](https://github.com/coreos/container-linux-config-transpiler/blob/master/doc/configuration.md) specification, some extra functionalities.

### Output

_Output_ defines the filename of the output for the file, if empty or not present the file is not written to the disk.

### Type

_Type_ defines the format of the output, the supported options are: `cloud-config`, `ignition` or `container-linux`, by default `container-linux` is used.

### Import

_Import_ option allows you to import another config file, merging the imported file over the _importing file_.

The _import_ key is a `map`, where the key is the *path* to the file to be imported, relative to the  _importing file_, the value is another `map` used to replaced using the [golang template system](https://golang.org/pkg/text/template/).

Given `foo.yaml`:

```yaml
---
import:
  bar.yaml:
    world: World!

foo: bar
baz: qux
```

And the `bar.yaml`:

```yaml
---
text: Hello {{.world}}
foo: qux
```

The result is:

```yaml
---
text: Hello World!
foo: qux
baz: qux
```

### Additional features

Additionally to the described features, a new schema is supported in `storage.file.content.remote.url`, the _file_ schema. When combustion is executed the file, relative to the yaml, is resolved and included inline.

Given `foo.yaml`

```yaml
---
storage:
  files:
    - path: foo.txt
      contents:
        remote:
          url: |
            file:///foo.txt
```

And the `foo.txt`

```
Hello World!
```

The result is:

```yaml
---
storage:
  files:
    - path: foo.txt
      contents:
        inline: Hello World!
``