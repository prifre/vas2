VAS - Visible Air System 

This is a program to use three different TSI instruments to measurem Air.
Two of the instruments, AeroTrak and DustTrak can be used unsupervised. 
P-Trak unfortunately needs to be supplied with isopropanol regulary so therefore it is not 
as flexible as the other two.

I was tasked to create a program where it would be possible to see in realtime how the air 
looks based on these three instruments. If the used does not have the instruments, there is a simple
simulation mode.

Then the project kind of grew based on various wishes, like being able to have continous measurement,
have a database with the measurements, export measurements, upload them to the Internet using ftp.
And eventually begoming a quite big and code-wise messy.

I used go and Fyne for the user interface. Other peoples go-packages are used for communication with instruments,
graphs, ftp, exports, etc.

With a simple
/Peter Freund
prifre@prifre.com
Homepage: prifre.com/vas
Download: prifre.com/vas/vas.zip