package asset_create

import (
	"encoding/json"
	"fmt"
	"github.com/dsoprea/go-exif"
	"io/ioutil"
	"os"
	"strings"
)

var (
	filepathArg     = ""
	printAsJsonArg  = false
	printLoggingArg = false
)

type IfdEntry struct {
	IfdPath     string                `json:"ifd_path"`
	FqIfdPath   string                `json:"fq_ifd_path"`
	IfdIndex    int                   `json:"ifd_index"`
	TagId       uint16                `json:"tag_id"`
	TagName     string                `json:"tag_name"`
	TagTypeId   exif.TagTypePrimitive `json:"tag_type_id"`
	TagTypeName string                `json:"tag_type_name"`
	UnitCount   uint32                `json:"unit_count"`
	Value       interface{}           `json:"value"`
	ValueString string                `json:"value_string"`
}

func PhotoHasExifData(filePath string) bool {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Open file: ", err)
		return false
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return false
	}

	_, err = exif.SearchAndExtractExif(data)
	if err != nil {
		//fmt.Println("SearchAndExtractExif: ", err)
		return false
	}
	return true
}

func ReadExifData(filePath string) (bool, string) {

	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Open file: ", err)
		return false, ""
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return false, ""
	}

	rawExif, err := exif.SearchAndExtractExif(data)
	if err != nil {
		fmt.Println("SearchAndExtractExif: ", err)
		return false, ""
	}

	// Run the parse.

	im := exif.NewIfdMappingWithStandard()
	ti := exif.NewTagIndex()

	entries := make([]IfdEntry, 0)

	visitor := func(fqIfdPath string, ifdIndex int, tagId uint16, tagType exif.TagType, valueContext exif.ValueContext) (err error) {

		ifdPath, err := im.StripPathPhraseIndices(fqIfdPath)
		if err != nil {
			fmt.Println("Panic 4: ", err)
			return err
		}

		it, err := ti.Get(ifdPath, tagId)
		if err != nil {
			if err == nil {
				fmt.Println("Panic 5: ", exif.ErrTagNotFound)
				fmt.Printf("WARNING: Unknown tag: [%s] (%04x)\n", ifdPath, tagId)
				return nil
			} else {
				fmt.Println("Panic 5: ", err)
			}

			//if log.Is(err, exif.ErrTagNotFound) {
			//	fmt.Printf("WARNING: Unknown tag: [%s] (%04x)\n", ifdPath, tagId)
			//	return nil
			//} else {
			//	//log.Panic(err)
			//}
		}

		valueString := ""
		var value interface{}
		if tagType.Type() == exif.TypeUndefined {
			var err error
			value, err = valueContext.Undefined()
			if err != nil {
				if err == exif.ErrUnhandledUnknownTypedTag {
					value = nil
				} else {
					fmt.Println("Panic 6: ", err)
					//log.Panic(err)
				}
			}

			valueString = fmt.Sprintf("%v", value)
		} else {
			valueString, err = valueContext.FormatFirst()
			//log.PanicIf(err)

			value = valueString
		}

		entry := IfdEntry{
			IfdPath:     ifdPath,
			FqIfdPath:   fqIfdPath,
			IfdIndex:    ifdIndex,
			TagId:       tagId,
			TagName:     it.Name,
			TagTypeId:   tagType.Type(),
			TagTypeName: tagType.Name(),
			UnitCount:   valueContext.UnitCount(),
			Value:       value,
			ValueString: valueString,
		}

		entries = append(entries, entry)

		return nil
	}

	_, err = exif.Visit(exif.IfdStandard, im, ti, rawExif, visitor)
	//log.PanicIf(err)
	if err != nil {
		fmt.Println("Panic 7: ", err)
	}

	if printAsJsonArg == true {
		data, err := json.MarshalIndent(entries, "", "    ")
		//log.PanicIf(err)
		fmt.Println("Panic 8: ", err)
		fmt.Println(string(data))
	} else {
		for _, entry := range entries {

			//fmt.Printf("IFD-PATH=[%s] ID=(0x%04x) NAME=[%s] COUNT=(%d) TYPE=[%s] VALUE=[%s]\n",
			//	entry.IfdPath,
			//	entry.TagId,
			//	entry.TagName,
			//	entry.UnitCount,
			//	entry.TagTypeName,
			//	entry.ValueString)

			if strings.Contains(entry.TagName, "Orientation") {
				//fmt.Println("description value: " + entry.ValueString)
				return true, entry.ValueString
			}
		}
	}

	return false, ""
}

func getSubstringAfterLastDot(s string) string {
	// Find the last occurrence of the dot
	lastDotIndex := strings.LastIndex(s, ".")

	// If there's no dot, return an empty string or the original string
	if lastDotIndex == -1 {
		return ""
	}

	// Return the substring after the last dot
	return s[lastDotIndex+1:]
}

func getSubstringBeforeLastDot(s string) string {
	// Find the last occurrence of the dot
	lastDotIndex := strings.LastIndex(s, ".")

	// If there's no dot, return the original string
	if lastDotIndex == -1 {
		return s
	}

	// Return the substring before the last dot
	return s[:lastDotIndex]
}
