# Minnow Tutorial: Readin Files

This is a tutorial for using the pre-alpha version of the minnow compression library. Later versions will make this easier.

Please inform me if you find errors in this tutorial at mansfield@uchicago.edu.

## Reading Minnow Files in Python

Most users will never need to generate minnow files. If you fall into this group and only want to read files in Python, installation is relatively simple.

1. Make sure that you have the `numpy` library installed.
2. Clone the git repo into a location of your choice using the command `https://github.com/phil-mansfield/minnow.git`.
3. Once this directory is located at `xxxx/yyyy/zzzz/minnow`, let Python know where the corresponding Python library is by adding the directory `xxxx/yyyy/zzzz/minnow/python` to the `PYTHONPATH` environment variable. If you've never done something like this before, look at the instructions [here](https://stackoverflow.com/questions/3402168/permanently-add-a-directory-to-pythonpath) for unix/Mac and [here](http://net-informations.com/python/intro/path.htm) for Windows.
4. Restart your terminal, if you are using one.

To read halo catalogues compressed with minnow you will need to use the `minh` library.  The following code block shows an example of how to use this library:
```python
import minh

f = minh.open("Bolshoi.minh")
x, y, z, mvir = f.read(["x", "y", "z", "mvir"])
f.close()
```
`minh` files also contain useful information about the file and the simulation it came from
```python
import minh

f = minh.open("Bolshoi.minh")

# Prints the number of haloes in the file.
print(f.length)

# Prints the header of the original text file used to generate
# the minh file.
print(f.text) 

# Prints the names of all the variables contained in the file.
print(f.names)

# Prints the length of one side of the simulation box in Mpc/h.
print(f.L)

f.close()
```

Variables in `minh` files are read as a single array by default. However, it may not be possible to load all the haloes in a large simulation simultaneously due to memory restrictions. In these cases, haloes can be loaded in "blocks". 

The code below shows an example of how to compute the geometric mean of halo masses in a file normal and with blocks.
```python
import minh

f = minh.open("Bolshoi.minh")

# Standard way to compute geometric mean:
mvir = f.read(["mvir"])
geo_mean = 10**(np.sum(np.log10(mvir)) / len(mvir))

# Computing the same quantity with blocks:
log_sum = 0.0
for b in f.blocks:
    mvir = f.block(b, ["mvir"])
    log_sum += np.sum(np.log10(mvir))

geo_mean = 10**(log_sum / f.length)
        
f.close()
```
       
The lengths of each block can be found in the array `f.block_lengths` .
        
Sometimes it is important for haloes to be in the same block as all their neighboring haloes. In some minnow files, called "boundary files,"  blocks correspond to cubic cells. These blocks will contain all the haloes within those cubes as well as all the haloes in a thin, shared "boundary" layer around that cube. By convention, these minnow files will be named `xxxxx.bnd.minh` instead of `xxxxx.minh`. However, this can also be checked by calling `f.is_boundary()` after a minh file, `f`, has been read.
        
For boundary files, the `minh` library refers to the data contained in the central cubic region as a "cell" and the data corresponding to the cell and its surrounding layer as a "block."
        
Several methods of `minh` file objects are only used for boundary files and boundary files have several additional fields:
```python
import minh
        
f = minh.read("Bolshoi.bnd.minh")
        
# Prints True if the file is a boundary file
print(f.is_boundary())
        
# Prints the width of each cide of a cubic cell or block.
print(f.cell_width())
print(f.block_width())
        
# Prints the origin of block 7 in Mpc/h.
print(g.cell_origin(7))
        
# Prints the number of cubic cells on each side of the simulation.
print(f.cells)
        
# Prints the width of the boundary layer around each cubic cells
# in Mpc/h.
 print(f.boundary)
```
        
Additionally, boundary files have an additional variable, `"bnd"`, which is True if a halo is a block's boundary layer and False if the halo is a block's cell. This can be read like any other variable:
        
```python
import minh 
        
f = minh.read("Bolshoi.bnd.minh")
x, y, z, mvir, bnd = minh.read(["x", "y", "z", "mvir", "bnd"])
        
f.close()
```
