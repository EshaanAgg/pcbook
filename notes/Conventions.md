# Protocol Message

- The name of the message is UpperCase while the field name is lower_snake_case.
- The tag associated with each field is a number, where the fields `1-15` take 1 byte and those from `16-2047` take 2 bytes. Thus is it best to give the lower tags to the most frequent fields.
- Tags need not be in order. They also must be unique for the same level fields.
