Output demo.gif

Type "curl -T helloworld.txt lanfile.example/api/u/name/helloworld.txt"
Sleep 500ms 

Enter 

Sleep 5s 

Type "curl lanfile.example/api/d/name/helloworld.txt"
Enter

Sleep 5s

Type "wget lanfile.example/api/d/name/encrypted_file.txt.gpg"


Sleep 500ms
Enter
Sleep 500ms

Type "gpg --batch --passphrase-file password.txt -o secrets.txt -d encrypted_file.txt.gpg"

Enter

Sleep 10s