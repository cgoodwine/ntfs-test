package main


import (
  "bytes"
  "fmt"
  "os"
	"encoding/json"
  "io/ioutil"
  "io"

  "www.velocidex.com/golang/go-ntfs/parser"
//  "www.velocidex.com/golang/go-ntfs"
)

func icat(filename string) {
  fmt.Printf("icat function with %s filename\n", filename);

  // command line arguments
  cat_command_image_offset := int64(0);

  buf, err := ioutil.ReadFile(filename)
  if err != nil {
    fmt.Printf("Unable to open file %s. Please make sure to provide the full path to an existing file\n", filename);
    return;
  } else {
    fmt.Printf("Succesfully opened file %s\n", filename);
  }

  cat_command_file_arg := bytes.NewReader(buf);
  cat_command_arg := "1";
  cat_command_offset := int64(0);

	reader, _ := parser.NewPagedReader(&parser.OffsetReader{
		Offset: cat_command_image_offset,
		Reader: getReader(cat_command_file_arg),
	}, 1024, 10000)

	ntfs_ctx, err := parser.GetNTFSContext(reader, 0)
	fmt.Printf("Can not open filesystem: %v", err)

	mft_entry, err := GetMFTEntry(ntfs_ctx, cat_command_arg)
	fmt.Printf("Can not open path: %v", err)

	var ads_name string = ""
	// Access by mft id (e.g. 1234-128-6)
	_, attr_type, attr_id, ads_name, err := parser.ParseMFTId(cat_command_arg)
	if err != nil {
		attr_type = 128 // $DATA
	}

	data, err := parser.OpenStream(ntfs_ctx, mft_entry,
		uint64(attr_type), uint16(attr_id), ads_name)
	fmt.Printf("Can not open stream: %v", err)

	var fd io.WriteCloser = os.Stdout

	buf2 := make([]byte, 1024*1024*10)
	offset := cat_command_offset
	for {
		n, _ := data.ReadAt(buf2, offset)
		if n == 0 {
			return
		}
		fd.Write(buf2[:n])
		offset += int64(n)
	}
  

  return;
}


func istat(filename string) {
  fmt.Printf("istat function with %s filename\n", filename);

  // command line args
  stat_command_image_offset := int64(0);
  stat_command_verbose := true;
  verbose_flag := true;

  buf, err := ioutil.ReadFile(filename)
  if err != nil {
    fmt.Printf("Unable to open file %s. Please make sure to provide the full path to an existing file\n", filename);
    return;
  } else {
    fmt.Printf("Succesfully opened file %s\n", filename);
  }

  stat_command_file_arg := bytes.NewReader(buf);
  stat_command_arg := "1";

  // start with false for now
  stat_command_i30 := false;

	reader, _ := parser.NewPagedReader(&parser.OffsetReader{
//		Offset: *stat_command_image_offset,
		Offset: stat_command_image_offset,
		Reader: getReader(stat_command_file_arg),
	}, 1024, 10000)

	ntfs_ctx, err := parser.GetNTFSContext(reader, 0)
	fmt.Printf("Can not open filesystem: %v", err)

	if stat_command_verbose {
		ntfs_ctx.SetOptions(parser.Options{
			IncludeShortNames: true,
			MaxLinks:          1000,
			MaxDirectoryDepth: 100,
		})
	}

	mft_entry, err := GetMFTEntry(ntfs_ctx, stat_command_arg)
	fmt.Printf("Can not open path: %v", err)

	if verbose_flag {
		fmt.Println(mft_entry.Display(ntfs_ctx))

	} else {
		stat, err := parser.ModelMFTEntry(ntfs_ctx, mft_entry)
		fmt.Printf("Can not open path: %v", err)

		serialized, err := json.MarshalIndent(stat, " ", " ")
		fmt.Printf("Marshal: %v", err)

		fmt.Println(string(serialized))
	}

	if stat_command_i30 {
		i30_list := parser.ExtractI30List(ntfs_ctx, mft_entry)
		fmt.Printf("Can not extract $I30: %v", err)

		serialized, err := json.MarshalIndent(i30_list, " ", " ")
		fmt.Printf("Marshal: %v", err)

		fmt.Println(string(serialized))
	}

  return;
}


func help() {
  fmt.Printf("Usage:\n./executable_name icat <image> <path>\n./executable_name istat <image> <path>\n");
}

func main() {
  // Input options:
  // ./read-file <image> <path>
  // ./myprogram icat <image> <path>
  // ./myprogram istat <image> <path>

  // input required is (icat|istat) (image) (path)

  var arguments = os.Args
  // Argument 0 is executable name
  if len(arguments) < 3 {
    help()
    return;
  }

  // Argument one is icat or istat
  var filename = arguments[2]
  if arguments[1] == "istat" {
    istat(filename);
  } else if arguments[1] == "icat" {
    icat(filename);
  } else {
    fmt.Printf("Error: must enter either icat or istat\n");
    return;
  }
}
