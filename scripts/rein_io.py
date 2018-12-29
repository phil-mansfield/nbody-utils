""" rein_io.py contains routines for reading the various file formats used by
the DMO data stored on rein.

Module contents:
    BinHFile   - object representing an open .binh file.
    BinHHeader - object representing the header information in a .binh file.

This file can also be run from the command line to print information about a
given file. Run the file with no arguments for a usage summary.
"""

from __future__ import print_function
import array
import struct
import numpy as np
import sys
import os.path

class BinHHeader(object):
    """ BinHHeader represents the header of a .binh file. It contains a number
    of internal, private fields and the following public fields:
        header_text  - the header text of the original text halo catalogue
        int_columns  - the indices of columns that are stored as int64s instead
                       of float32s
        column_names - list of names for each column, in order
    """
    pass

class BinHFile(object):
    """ BinHFile represents an open .binh file.
    
    Example usage:

        binh = BinHFile("example.binh")
        header = binh.read_header()
        
        # Columns can be specified as an index.
        col_2, col_15, col_5 = binh.read_columns([2, 15, 5])
        
        # They can also be specified by their name (check bhf.column_names).
        m_vir, x, y, z = binh.read_columns(["MVir", "X", "Y", "Z"])
        
        # Use to following to read everything in the file.
        list_of_columns = binh.read_columns(header.column_names)
    """
    def __init__(self, fname):
        self.fname = fname
        self._hd = BinHHeader()

        fp = open(self.fname, "rb")
        int_dtype = np.dtype(np.int64).newbyteorder("<")
        
        # Read binary header

        bin_h_header_text = fp.read(7*8)

        (self._hd._mode,
         self._hd._haloes, self._hd._column_num,
         self._hd._mass_column, self._hd._int_column_num,
         self._hd._names_len, self._hd._text_header_len) = struct.unpack(
             "<" + "q"*7, bin_h_header_text
        )

        # Read int columns

        int_column_bytes = fp.read(8*self._hd._int_column_num)
        self._hd.int_columns = np.frombuffer(
            int_column_bytes, dtype=int_dtype
        )

        # Read column names

        column_names_bytes = fp.read(self._hd._names_len)
        self._hd.column_names = column_names_bytes.split("\n")
        self._hd.text_header = fp.read(self._hd._text_header_len)

        fp.close()

        # Setup lookup tables

        self._name_lookup = { }
        for i in xrange(len(self._hd.column_names)):
            self._name_lookup[self._hd.column_names[i]] = i

        self._int_lookup = { }
        for col in self._hd.int_columns:
            self._int_lookup[col] = True

    ####################
    ## public methods ##
    ####################

    def read_header(self):
        """ read_header() returns the .binh file's header as a BinHHeader. """
        return self._hd
        
    def read_columns(self, columns):
        """ read_columns(columns) reads and the given list of columns and
        returns them as numpy arrays with types of either np.int64 or
        np.float32. Columns can be specified by either an integer giving the
        column index or by a string giving the column name.
        """
        fp = open(self.fname, "rb")

        out = [None]*len(columns)

        int_dtype = np.dtype(np.int64).newbyteorder("<")
        float_dtype = np.dtype(np.float32).newbyteorder("<")

        for i in xrange(len(columns)):
            col = self._column_index(columns[i])
            if col in self._int_lookup:
                offset = self._int_offset(col)
                fp.seek(offset, 0)
                bytes = fp.read(8*self._hd._haloes)
                out[i] = np.frombuffer(bytes, dtype=int_dtype)
            else:
                offset = self._float_offset(col)
                fp.seek(offset, 0)
                bytes = fp.read(4*self._hd._haloes)
                out[i] = np.frombuffer(bytes, dtype=float_dtype)
            
        fp.close()
        
        return out

    #####################
    ## private methods ##
    #####################

    def _column_index(self, column_id):
        if type(column_id) == str:
            if column_id in self._name_lookup:
                return self._name_lookup[column_id]
            else:
                raise ValueError("Column name '%s' not recognized" % column_id)
        elif column_id >= self._hd._column_num:
            raise ValueError("Column %d out of range: max column number is %d" %
                             (column_id, self._hd._column_num))
        else:
            return column_id

    def _int_offset(self, col):
        # TODO: vecorize
        idx = -1
        for i in xrange(len(self._hd.int_columns)):
            if self._hd.int_columns[i] == col:
                idx = i
                break
        if idx == -1: raise ValueError("Impossible")

        offset = (8*7 + # Binary header
                8*self._hd._int_column_num + # Int columns
                self._hd._names_len + # Columns names
                self._hd._text_header_len + # Text header
                idx*8*self._hd._haloes) # Other int columns

        return offset

    def _float_offset(self, col):
        # TODO: vecorize
        idx = col
        for i in xrange(len(self._hd.int_columns)):
            if self._hd.int_columns[i] < col: idx -= 1
            if self._hd.int_columns[i] == col: ValueError("Impossible")

        offset = (8*7 + # Binary header
                8*self._hd._int_column_num + # Int columns
                self._hd._names_len + # Columns names
                self._hd._text_header_len + # Text header
                8*self._hd._int_column_num*self._hd._haloes + # Int columns
                4*idx*self._hd._haloes) # Other float columns

        return offset

def main():
    if len(sys.argv) != 2:
        print("""Proper command line usage:
    $ python rein_io.py file.binh - Prints the original text header of the halo
                                    catalogue.

Other uses to follow.
""")
        exit(1)

    fname = sys.argv[1]
    extension = os.path.basename(fname).split(".")[-1]

    if extension == "binh":
        bhf = BinHFile(fname)
        header = bhf.read_header()
        print(header.text_header, end="")
    else:
        print("""%s has an unrecognized file extension.
Currently, only .binh files are recognized.""")
        exit(1)


if __name__ == "__main__": main()
