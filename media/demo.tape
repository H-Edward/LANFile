Output demo.gif

Type "curl -T helloworld.txt lanfile.example/u/name/helloworld.txt"
Sleep 500ms 

Enter 

Sleep 5s 

Type "curl lanfile.example/d/name/helloworld.txt"
Enter

Sleep 5s

Type "gpg --encrypt --recipient edward important_document.pdf"
Enter

Sleep 500ms

Type "curl -T important_document.pdf.gpg lanfile.example/u/name/encrypted_document.pdf.gpg?encrypted=key""

Sleep 500ms

Enter

Sleep 10s