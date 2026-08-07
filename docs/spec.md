# Byte Layout

- byte layout we are going to used here
  - 1 byte Type - 4 bytes Length - Length bytes Payload -
  - possible values for Type field
    - 0x01 - Data
    - 0x02 - heartbeat
    - 0x03 - Close

- Maximum Frame Size : 64\*1024 bytes

- decode algorithm we are using
  - Read first byte for type
  - Read next 4 bytes for length
  - Validate that length <= Maximum Frame Size
  - Read exactly Lengh bytes for payload
