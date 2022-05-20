# based

Let's set up a simple dynamic form example.
First, create the YAML file that will contain the form components and layout. We'll replicate the following example:

![(Text) Sign up!, (Text input) Your name, (Text input) Your email address, (Check box) I agree with the Terms and conditions, (Button) Next](/readme_example.png?raw=true)

Our `sign-up.yml` file should look like this:

```yml
_version: 1
title:
  _item: title
  text: signUp
name:
  _item: textbox
  _type: field
  _rules:
    - _action: regex
      _param: ^[A-Z]+\w*(\s\w+)*$
      invalidMsg: invalidFullName
  text: yourFullName
email:
  _item: textbox
  _type: field
  _rules:
    - _action: regex
      _param: ^\w+([\.-]?\w+)*@\w+([\.-]?\w+)*(\.\w{2,3})+$
      invalidMsg: invalidEmail
  text: signUp
password1:
  _item: textbox
  _type: field
  _rules:
    - _action: regex
      _param: ^[A-Z]+\w*(\s\w+)*$
      invalidMsg: invalidPassword
  text: yourPassword
  hidden: true
password2:
  _item: textbox
  _rules:
    - _action: ==
      _param: password1
      invalidMsg: passwordsDontMatch
  text: repeatYourPassword
  hidden: true
confirm:
  _item: button
  _type: action
  text: next
```

Here are some rules to follow when writing YAML files for your forms:

- The name of the file indicates the ID of the form;
- Each item has a key that indicates its own ID;
- **based** allows you to insert custom keys anywhere (such as `invalidMsg` and `text` from the example above). They will be forwarded to the compiled version of the form;
- It is a good practice to, rather than type the literal text that the form should display to the user, implement a string localization system on the front end and use the string IDs instead;
- The keys that start with an underscore (such as `_item`) are reserved and will not be forwarded to the compiled form;
- `_item` indicates the component the front end must render for this item;
- `_type` can have 3 values: `none` (default if ommited), `field` (that indicates the front end should send data through this item), and `action` (that indicates the front end is able to perform an action through this item, such as confirming or canceling some process);
- `_rules` is a list of rules that the value of the field must follow. If the current item has rules but is not a field, then they will not be validated by **based** (in those cases, it's assumed that the front end is responsible for those validations);
- A rule has an `_action` (that can be `==` for equality, `!=` for inequality, and `regex`, for regular expressions) and a `_param` (that, in case of equality/inequality operations, indicates which item ID to be compared to; and in case of a regex operation, the regular expression it should be matched to).

Once compiled, the form's DTO will look like this:

```json
{
  "name": "sign-up",
  "actions": ["confirm"],
  "fields": ["name", "email", "password1"],
  "layout": [
    {
      "id": "name",
      "item": "textbox",
      "props": {
        "text": "yourFullName"
      },
      "rules": [
        {
          "action": "regex",
          "param": "^[A-Z]+\\w*(\\s\\w+)*$",
          "props": {
            "invalidMsg": "invalidFullName"
          }
        }
      ]
    },
    {
      "id": "email",
      "item": "textbox",
      "props": {
        "text": "signUp"
      },
      "rules": [
        {
          "action": "regex",
          "param": "^\\w+([\\.-]?\\w+)*@\\w+([\\.-]?\\w+)*(\\.\\w{2,3})+$",
          "props": {
            "invalidMsg": "invalidEmail"
          }
        }
      ]
    },
    {
      "id": "password1",
      "item": "textbox",
      "props": {
        "hidden": true,
        "text": "yourPassword"
      },
      "rules": [
        {
          "action": "regex",
          "param": "^[A-Z]+\\w*(\\s\\w+)*$",
          "props": {
            "invalidMsg": "invalidPassword"
          }
        }
      ]
    },
    {
      "id": "confirm",
      "item": "button",
      "props": {
        "text": "next"
      }
    }
  ]
}
```
