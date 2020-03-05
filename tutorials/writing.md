# Minnow Tutorial: Writing Files

This is a tutorial for using the pre-alpha version of the minnow compression library. Later versions will make this easier.

Please inform me if you find errors in this tutorial at mansfield@uchicago.edu.

## Writing Minnow Files in Go

Currently, minnow files can only be written in one language: the C-like language called "Go." 

To download a Go compiler, use [this](https://golang.org/dl/) website. Follow the instructions [here](https://golang.org/doc/install) to install. The instructions are relatively painless as far as compilers go. Let me know if you run into significant problems and I will try to help.

To get my code, type  
`$ go install github.com/phil-mansfield/minnow` 
from anywhere. If this fails, tell me what happened, but as a back up, you can manually clone the minnow directory into
`$HOME/go/src/github.com/phil-mansfield/minnow`
You will need to create some of those directories first. In that case, type
`$ go install`
while in the minnow directory.

To test that things are working, go to  
`$HOME/go/src/github.com/phil-mansfield/minnow/go/minh`  
and type  
`$ go test`  
If anything goes wrong, let me know.
  
If all goes well you're ready to run the compressor. go to  
`$HOME/go/src/github.com/phil-mansfield/minnow/scripts`  
and run  
`$ go build text_to_minh.go`  
which will make a binary called `text_to_minh`. This binary will convert text halo catalogues to minh files. This is what a configuration file for [BolshoiP](https://www.cosmosim.org/cms/simulations/bolshoip/) looks like: 
``` 
[minh]
L = 250
Epsilon = 1
Mp = 1.54975e+08
MinParticles = 50
NameIndex = name_index.txt
TypeIndex = full_type_index.txt
```
This is what it looks like for [MDPL2](https://www.cosmosim.org/cms/simulations/mdpl2/)
```
[minh]
L = 1000
Epsilon = 5 
Mp = 1.50466e+09  
MinParticles = 50  
NameIndex = name_index.txt  
TypeIndex = full_type_index.txt
```
L is the width in Mpc/h, Epsilon is the force softening scale in kpc/h, Mp is the particle mass, and the last two lines reference helper files that specify how the file should be compressed and how to parse the halo file. If you are converting Rockstar files to the minh format, the `name_index.txt` and `full_type_index.txt` files will work regardless of version and will convert the files into a simple binary format.
  
Supposing that you have a config file called `BolshoiP.config`, you can convert files using the `text_to_minh` binary which you created earlier. This can be done with:
`$ ./text_to_minh BolshioP.config all hlist_1.00000.list output_directory_name`  
or  
`$ ./text_to_minh BolshioP.config all "*.list" output_directory_name`  
if you have a lot of BolshoiP halo files.
