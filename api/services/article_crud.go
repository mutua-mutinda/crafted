package services

import (
	"strconv"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/vonmutinda/crafted/api/channels"
	"github.com/vonmutinda/crafted/api/database"
	"github.com/vonmutinda/crafted/api/messages"
	"github.com/vonmutinda/crafted/api/models"
)

// ArticleService struct
type ArticleService struct {
	Logger 	*logrus.Logger
	DB 	*gorm.DB
}
 
// GetAllArticles returns all articles
func (a *ArticleService) GetAllArticles() ([]models.Article, error){
	var err error
 
	articles := []models.Article{}
 
	done := make(chan bool)

	// go routine to fetch articles
	go func(c chan<- bool){   
		if err = a.DB.Preload("Author").Find(&articles).Error ; err != nil {
			a.Logger.Errorf("cannot fetch articles: %v", err)
		}
		c<- true 
	}(done) 

	if channels.OK(done){ 
		return articles,nil 
	}
	return nil, err
}

// SaveArticle func
func (a *ArticleService) SaveArticle(article *models.Article) (*models.Article, error){
	var err error 
	
	done := make(chan bool)
	go func(c chan<- bool){
		err = a.DB.Model(&models.Article{}).Create(&article).Error 
		if err != nil {  
			c<- false
			return
		} 
		
		err = database.GetDB().Where("id = ?", article.AuthorID).Take(&article.Author).Error
		if err != nil {
			a.Logger.Errorf("cannot fetch article's author id %d : %v", article.AuthorID)
			c<- false
		}
		c<- true 
	}(done)

	if channels.OK(done){ 
		return article,nil 
	}
	return &models.Article{}, err
}

// DeleteAllArticles func 
func (a *ArticleService) DeleteAllArticles() (int64, error){
	// var err error 
	var rep *gorm.DB
	done := make(chan int, 1)

	go func(c chan<- int){  
		rep = database.GetDB().Raw(`
			UPDATE articles
			SET deleted_at=? 
		`, time.Now())
		c<- 1
	}(done)
	
	<-done

	return rep.RowsAffected, rep.Error
}

// FetchArticleByID func
func (a *ArticleService) FetchArticleByID(id uint64) (models.Article, error){ 

	var err error 
	article := models.Article{}  
	done := make(chan bool)

	go func(c chan<- bool){
		if err = a.DB.Preload("Author").Where("ID = ?", id).Take(&article).Error; err != nil {
			a.Logger.Errorf("cannot fetch article by id %d : %v", id, err)
			c<- false
			return
		} 
		c<- true
	}(done)

	if channels.OK(done){ 
		return article, nil
	}

	if gorm.IsRecordNotFoundError(err){ 
		return models.Article{}, err
	}

	return models.Article{}, err	
}

// DeleteByID func
func (a *ArticleService) DeleteByID(id uint64) error { 
	 
	var err error  
	done := make(chan int, 1) // buffered channel

	go func(c chan<- int){ 
		err = a.DB.Exec(`
			DELETE FROM articles 
			WHERE id = ?
		`, id).Error

		if err != nil {
			a.Logger.Errorf("cannot delete article id : %d", id)
		}

		c<- 1
	}(done)

	<-done 
	return err
}


// UpdateArticle func 
// using wait groups for fun !!!! 
func(a *ArticleService) UpdateArticle(updated *models.Article, aid int64) (*models.Article, error) {

	var wg sync.WaitGroup  
	var err error

	wg.Add(1)
	go func(done *sync.WaitGroup){

		defer done.Done()   
		// for testing purpose let's delegate updating time to rabbitmq
		gor := a.DB.Exec(`
				UPDATE articles
				SET title=?,
					body=?
				WHERE id=?
			`,updated.Title,
			updated.Body,
			aid,
		) 

		if err =  gor.Error; err != nil {
			a.Logger.Errorf("cannot update article id %d : %v", aid, err)  
		}
 
	}(&wg)

	wg.Wait() 
	
	// send to queue -- for fun again!!! 
	s := strconv.FormatInt(aid, 10)
	messages.SendMessage("updated_at", s)

	return updated, err
} 