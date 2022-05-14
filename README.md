# based

## Getting started

Let's set up a simple dynamic form example.
First, we create two JSON files: `components.json` will contain the definition of components that we'll use on our forms, and `forms.json` will contain the forms that will be populated by the components.

In this example, we'll replicate the following form:

![(Text) Sign up!, (Text input) Your name, (Text input) Your email address, (Check box) I agree with the Terms and conditions, (Button) Next](https://i.imgur.com/c03EzLS.png)

Our `components.json` file should look like this:

```jsonc
{
  "__version": 0,

  "title": {
    "text": "string"
  },
  "textbox": {
    "text": "string",
    "hint": "string"
  },
  "checkbox": {
    "text": "string",
    "required": "bool"
  },
  "button": {
    "text": "string",
    "enabled": {
      "__type": "bool",
      "__default": true
    }
  }
}
```

As you can see, we have successfully defined the field types that we'll use on our form. Now, to the `forms.json`file:

```jsonc
{
  "__version": 0,

  "signUp": {
    "signUpTitle": {
      "__type": "title",
      "text": "Sign Up!"
    },
    "nameInput": {
      "__type": "textbox",
      "hint": "Your Name"
    },
    "emailInput": {
      "__type": "textbox",
      "hint": "Your Email"
    },
    "termsCheckbox": {
      "__type": "checkbox",
      "text": "I agree with the Terms and conditions",
      "required": true
    },
    "nextButton": {
      "__type": "button",
      "text": "Next"
    }
  }
}
```

## Components Definition File Guide

Components should be defined on a JSON file. This file has to be an object with the `__version` key defined. Its value should be an integer, which represents the Components Definition File version. The rest of the keys of the root object should be the names of the components, with their values being the components structure. Example:

```jsonc
{
    "__version": 0,

    "myTextBox": {...},
    "myButton": {...}
}
```

The structure of a component is an object whose keys are the names of the fields of that component. Their values can be a string, for simple typing, or an object, for advanced typing.

If a field is defined by simple typing, the string must be the name of one of the primitive types (`string`, `number`, and `bool`). Example:

```jsonc
{
  "__version": 0,

  "image": {
    "url": "string" // Primitive type
  },

  "button": {
    "text": "string", // Primitive type
    "icon": "image" // Reference to previous component
  }
}
```

If you want to define a default value for that field, for example, you could use advanced typing. In that case, the value must be an object with the `__type` key defined, and its value should be a string with the name of a primitive type (`string`, `number`, or `bool`). Then, you could specify a `__default` key on that object with the desired value.

Examples of definition by advanced typing:

```jsonc
{
  "__version": 0,

  "image": {
    "url": "string"
  },

  "button": {
    // Advanced typing
    "text": {
      "__type": "string",
      "__default": "Ok"
    }
  }
}
```

You can use components inheritance to create variations of a given component. Just add the `__inherit` key, whose value is a string: the name of the component that you want to inherit the fields from. Defining a field that already exists on the inherited component will override it.

Components can also have an `__abstract` key with a boolean value, which represents if that component is meant to only serve as a template for others, and cannot be directly implemented on forms. Example:

```jsonc
{
  "__version": 0,

  "disableableComponent": {
    "__abstract": true, // This is an abstract component, therefore it cannot be directly implemented
    "disabled": "bool"
  },

  "button": {
    "__inherit": "disableableComponent", // The "button" component inherits the fields from the previous component
    "text": "string",
    "color": {
      "__type": "string",
      "__default": "#fff"
    }
  },

  "submitButton": {
    "__inherit": "button",
    "color": {
      "__type": "string",
      "__default": "#00ff00" // This "button" variant overrides the "color" field to change its default value
    }
  }
}
```

> The order of the definition of components matter. That means you can only reference a component _after_ it is defined on this file.

## Forms Definition File Guide

Forms should be defined on a JSON file. This file has to be an object with the `__version` key defined. Its value should be an integer, which represents the Components Definition File version. The rest of the keys of the root object should be the names of the forms, with their values being the forms structure. Example:

```jsonc
{
    "__version": 0,

    "signUpForm": {...},
    "logInForm": {...}
}
```

The structure of a form is an object whose keys are the names of the component instances that the form will contain. They should be JSON objects with a `__type` key defined, whose value should be the name of the component that will be instanced. The rest of the keys inside of the instance should be names of the fields of that component, followed by their values. You should specify the value of every field present on the component definition, except by fields with default value, which are optional. Example:

```jsonc
// components.json
{
  "__version": 0,

  "label": {
    "text": "string"
  },

  "button": {
    "text": "string",
    "enabled": {
      "__type": "bool",
      "__default": true
    }
  }
}
```

```jsonc
// forms.json
{
  "__version": 0,

  "notice": {
    "paragraph": {
      "__type": "label",
      "text": "Notice: This is to inform that your time has come."
    },
    "confirmButton": {
      "__type": "button",
      "text": "Okay"
    }
  }
}
```
