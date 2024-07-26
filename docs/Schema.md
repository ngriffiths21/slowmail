# Database Schema

### Mail table

- `user_id` (integer): Slow Mail user ID
- `orig_date` (integer): Date time received in mail header, in Unix seconds
- `date` (integer): Date of delivery, rounded to midnight and given in Unix seconds
- `from_head` (string): Combined content of from, sender, and reply-to mail headers
- `from_name` (string): Display name of primary sender
- `from_addr` (string): Email of primary sender
- `to_head` (string): Combined content of to and cc mail headers
- `message_id` (string): Message ID from mail header
- `in_reply_to` (string): Content of in-reply-to header, with message IDs of parent message(s)
- `subject` (string): Content of subject header
- `content` (string): Mail body
- `multifrom` (integer): Boolean flag for more than one from address
- `multito` (integer): Boolean flag for more than one to address

### User table

- `user_id`: Slow Mail user ID
- `username`: Email username
- `password`: Hash of password
- `display_name`: Display name
- `recovery_addr`: Recovery email (optional)
