# goreport

Report handler to process ITSM data into country specific reports

# usage

goreport [options] command [noun]

### commands:
- report
- list countries
- list prodcategories
- list services

### options:

#### -v
Increase verbosity, provide some output as to what the program is doing.

#### -cfg `<filename>`
Load the configuraiton file specified by the filename. Defaults to 
`goreport.yaml` in the current directory.

#### -country `<country>`
Defines the country to use for loading the incidents. Defaults to the 
`defaultcountry` as defined in the configuration file.

#### -input `<filename>`
Defines the input tab-delimited file to load incidents from. This is a
UTF-16 file coming out of the data warehouse. It default to `allincidents.csv`.

#### -now
Useful in combination with the `report` command. Instead of reporting on last
month, it reports on the current month. 

#### -month `<int>`
Run a report on a specific month. Jan equals to 1, Dec to 12.

#### -year `<int>`
The year (in 4 digit format) to run the report on. Bear in mind that the
5 months preceding month need to have incidents in the input file.

#### -output `<filename>`
Specify the output filename. It will default to 
`report-<country>-<month>-<year>.xlsx` where `<month>` and `<year>` are 
numerical. Example: `report-sweden-10-2019.xlsx`.

#### -reference `<filename> | "same"`
Use a reference file to load updates form (excluded incidents and updated 
resolution times). If the string `same` is provided, it will use the default
filename (see `-output`) as the input file and update it.

#### -nofilter
Don't filter out any product categories that are defined in the configuration
file. 

#### -reverse
Use the product category filter in reverse (i.e. show what has been filtered 
out). 



