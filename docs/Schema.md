# Database Schema

### Mail table

- `user_id` (integer): Slow Mail user ID
- `orig_date` (unsigned int): Date time received in mail header, in Unix seconds
- `date` (unsigned int): Date of delivery, rounded to midnight and given in Unix seconds
- `from_head` (text): Combined content of from, sender, and reply-to mail headers
- `from_name` (varchar(40)): Display name of primary sender
- `from_addr` (varchar(255)): Email of primary sender
- `to_head` (text): Combined content of to and cc mail headers
- `message_id` (text): Message ID from mail header
- `in_reply_to` (text): Content of in-reply-to header, with message IDs of parent message(s)
- `subject` (text): Content of subject header
- `content` (text): Mail body
- `multifrom` (tinyint): Boolean flag for more than one from address
- `multito` (tinyint): Boolean flag for more than one to address

### User table

- `user_id` (integer primary key): Slow Mail user ID
- `username` (varchar(40)): Email username
- `password` (binary(64)): Hash of password
- `display_name` (varchar(40)): Display name
- `recovery_addr` (varchar(255)): Recovery email (optional)
