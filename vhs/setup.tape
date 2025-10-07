Output docs/demo.setup.gif

Set Shell "bash"
Set FontSize 18
Set Width 850
Set Height 680
Set Padding 16
Set Framerate 15
Set TypingSpeed 30ms

Type "./ai-cli setup"
Sleep 500ms
Enter
Wait+Screen@5000ms /No inference detected/

Enter
Wait+Screen@5000ms /Paste your API key below/

Type aValidApiKey
Enter

Sleep 5000ms

Type "./ai-cli chat"
Enter
Wait+Screen@5000ms /Welcome to the AI CLI!/
Type "who are you?"
Enter
Sleep 5000ms
