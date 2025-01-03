package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	elastic "gopkg.in/olivere/elastic.v5"
)

var (
	ErrNotFound = errors.New("entity not found")
)

type Repo interface {
	Close()
	PutProduct(ctx context.Context, p Product) error
	GetProductByID(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error)
	ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error)
}

type elasticRepo struct {
	client *elastic.Client
}

type productDocument struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func NewElasticRepo(url string) (Repo, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(url),
		elastic.SetSniff(false),
	)
	if err != nil {
		return nil, err
	}
	return &elasticRepo{client}, nil
}

func (r *elasticRepo) Close() {
}

func (r *elasticRepo) PutProduct(ctx context.Context, p Product) error {
	_, err := r.client.Index(). //Initiates an indexing operation in Elasticsearch using the client provided by the elastic library
					Index("catalog").         //specifies the name of the Elasticsearch index where the product will be stored.
					Type("product").          //Sets the document type within the index.
					Id(p.ID).                 //Sets the unique identifier (ID) for the document.
					BodyJson(productDocument{ //Specifies the document to be indexed in the Elasticsearch database.
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		}).Do(ctx) //Executes the indexing operation within the given context
	return err
}

func (r *elasticRepo) GetProductByID(ctx context.Context, id string) (*Product, error) {
	res, err := r.client.Get().Index("catalog").Type("product").Id(id).Do(ctx)
	if err != nil {
		return nil, err
	}
	if !res.Found {
		return nil, ErrNotFound
	}
	p := productDocument{}
	if err = json.Unmarshal(*res.Source, &p); err != nil { //Deserializes the JSON source into a productDocument struct
		return nil, err
	}
	return &Product{
		ID:          id,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}, err
}

func (r *elasticRepo) ListProducts(ctx context.Context, skip, take uint64) ([]Product, error) {
	res, err := r.client.Search(). //Starts a search operation in Elasticsearch
					Index("catalog").                  //Specifies the index (catalog) where the documents (products) reside
					Type("product").                   //Specifies the type of document (product)
					Query(elastic.NewMatchAllQuery()). //Uses the MatchAllQuery to retrieve all documents from the catalog index
					From(int(skip)).Size(int(take)).
					Do(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	products := []Product{}
	for _, hit := range res.Hits.Hits { //The Hits field in res contains the list of documents that matched the query
		p := productDocument{}                                 //Hits is an array where each element is a hit representing a document
		if err = json.Unmarshal(*hit.Source, &p); err == nil { //json.Unmarshal deserializes this JSON into a productDocument struct p
			products = append(products, Product{
				ID:          hit.Id,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			})
		}
	}
	return products, err
}

func (r *elasticRepo) ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error) {
	items := []*elastic.MultiGetItem{}
	for _, id := range ids {
		items = append(
			items,
			elastic.NewMultiGetItem(). //For each id, a new MultiGetItem is created using elastic.NewMultiGetItem()
							Index("catalog").
							Type("product").
							Id(id), //Each id is added to the MultiGetItem
		)
	}
	res, err := r.client.MultiGet(). //The MultiGet API is used to retrieve multiple documents in a single request
						Add(items...). //Adds the items slice to the MultiGet request. Each item represents a product ID to retrieve.
						Do(ctx)        //Executes the MultiGet request asynchronously, using the provided ctx context
	if err != nil {
		log.Println(err)
		return nil, err
	}
	products := []Product{}
	for _, doc := range res.Docs {
		p := productDocument{}
		if err = json.Unmarshal(*doc.Source, &p); err == nil {
			products = append(products, Product{
				ID:          doc.Id,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			})
		}
	}
	return products, nil
}

func (r *elasticRepo) SearchProducts(ctx context.Context, query string, skip, take uint64) ([]Product, error) {
	res, err := r.client.Search().
		Index("catalog").
		Type("product").
		Query(elastic.NewMultiMatchQuery(query, "name", "description")).
		From(int(skip)).Size(int(take)).
		Do(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	products := []Product{}
	for _, hit := range res.Hits.Hits {
		p := productDocument{}
		if err = json.Unmarshal(*hit.Source, &p); err == nil {
			products = append(products, Product{
				ID:          hit.Id,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			})
		}
	}
	return products, err
}
