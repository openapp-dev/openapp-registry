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

### string
```
<parameter_name>:
    description: "<parameter's description>"
    type: string
    required: true
    default: "<default value>"
```

### boolean
```
<parameter_name>:
    description: "<parameter's description>"
    type: boolean
    required: true
    default: "<default value>"
```

### array
```
<parameter_name>:
    description: "<parameter's description>"
    type: array
    required: false
    itemProperties:
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
