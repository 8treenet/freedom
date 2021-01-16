# Contributing conventions

## Code style

### Comment

Limit all lines to around 80 characters.

Example:

1. Good.

   Example:

   ```go
   package main
   
   // Hello, Li. I've released a new version of the project, would you like to make
   // code view for it?
   ```

   It is exactly 80 characters in the first line.


2. Good.

   Example:

   ```go
   package main
   
   // Let's have a good dinner this Friday, Li? I heard a new restaurant is open from
   // my friend.
   ```

   It is 82 characters in the first line, but it is acceptable.

   It is NOT REQUIRED that you MUST limit each line <= 80 characters.

   For example, you've been typed 74 characters until "is", and your next word is "open". If it is required each line
   must <= 80 characters, you may not discover that it has been exceeded 80 characters before "open" has typed,
   therefore you must back to the start position of "open" and insert a newline to content the rule. So, it will lead to
   a bad experience in coding if it required each line MUST <= 80 characters.



3. Bad.

   ```go
   package main
   
   // I didn't know what I want to say, emmmm... Write some words here to make up. Aha, 
   // it has been finished.
   ```

   It is 84 characters in the first line (includes punctuations), but it is not acceptable.

   Why is it? After the word "make up" you've typed, you would directly know from the editor that it is 78 characters
   you've typed in this line. Whatever the next word is, you should realize you MUST insert a newline here, because of
   no words that consist of 1 character only. Even if you want to type a letter in the next position, you should also
   insert a newline for tidy. So, remember, whatever the next word is, if what you've typed is closed to 80 characters,
   insert a newline here, just for tidy.