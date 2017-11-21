package main

// #include "netcdf.h"
// #cgo LDFLAGS: -lnetcdf
import "C"

import (
	"fmt"
	"unsafe"
)


func main() {
	ncid := C.int(0)
	C.nc_open(C.CString("/g/data2/rs0/tiles/EPSG3577/LS5_TM_NBAR/LS5_TM_NBAR_3577_-3_-22_2006.nc"), C.NC_NOWRITE, &ncid)
	fmt.Println(ncid)
	
	varid := C.int(0)
	status := C.nc_inq_varid(ncid, C.CString("band_5"), &varid);
	fmt.Println(status, varid)

	data := make([]int16, 4000*4000)
	/* Read the data. */
        retval := C.nc_get_var_short(ncid, varid, (*C.short)(unsafe.Pointer(&data[0])))
	fmt.Println(retval)
}
