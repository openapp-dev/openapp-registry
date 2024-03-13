# OpenAPP Official Application Registry

## Inputs

It clould be the following formats:

### integer
```
<parameter_name>:
    description: "<parameter's description>"
    type: integer
    required: true
    default: 4443
```

Demo:
```yaml
para_a: 4443
```

### string
```
<parameter_name>:
    description: "<parameter's description>"
    type: string
    required: true
    default: "<default value>"
```


Demo:
```yaml
para_a: value_1
```

### boolean
```
<parameter_name>:
    description: "<parameter's description>"
    type: boolean
    required: true
    default: "<default value>"
```

Demo:
```yaml
para_a: false
```

### array
```
<parameter_name>:
    description: "<parameter's description>"
    type: array
    required: false
    items:
        type: object
        properties:
            <sub_paramater_1>:
                description: "<parameter's description>"
                type: int
                required: false
                defualt: 0
            <sub_paramater_2>:
                description: "<parameter's description>"
                type: boolean
                required: false
                defualt: false
```

Demo:
```yaml
para_a:
- sub_para_1: 1
  sub_para_2: true
- sub_para_1: 2
  sub_para_2: false
```

Or
```
<parameter_name>:
    description: "<parameter's description>"
    type: array
    required: false
    items:
        type: integer
```

Demo:
```yaml
para_a:
- 1
- 2
- 3
```

### object
```
<parameter_name>:
    description: "<parameter's description>"
    type: object
    required: true
    properties:
        <sub_paramater_1>:
            description: "<parameter's description>"
            type: int
            required: false
            defualt: 0
        <sub_paramater_2>:
            description: "<parameter's description>"
            type: boolean
            required: false
            defualt: false
```

Demo:
```yaml
para_a:
  sub_para_1: 1
  sub_para_2: true
```