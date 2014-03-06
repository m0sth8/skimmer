package skimmer

type Storage interface {
	LookupBin(name string) (*Bin, error) // get one bin element by name
	LookupBins(names []string) ([]*Bin, error) // get slice of bin elements
	LookupRequest(binName, id string) (*Request, error) // get request from bin by id
	LookupRequests(binName string, from, to int) ([]*Request, error) // get slice of requests from bin by position
	CreateBin(bin *Bin) error // create bin in memory storage
	UpdateBin(bin *Bin) error // save
	CreateRequest(bin *Bin, req *Request) error
}

type BaseStorage struct {
	maxRequests       int
	binLifetime		  int64
}
