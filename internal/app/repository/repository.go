package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Cost struct {
	ID    int
	Title string
	Img   string
	Info  string
	Type  bool
}

type Request struct {
	ID      int
	Cost_id int
	Price   float64
	Main    bool
}

type RequestView struct {
	ID         int
	Min_volume int
	Max_volume int
	Ratio      float64
}

func (r *Repository) GetRequestView() ([]RequestView, error) {
	requestView := []RequestView{
		{
			ID:         1,
			Min_volume: 500,
			Max_volume: 1000,
			Ratio:      1.01,
		},
	}
	if len(requestView) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}
	return requestView, nil
}

func (r *Repository) GetRequest() ([]Request, error) {
	request := []Request{
		{
			ID:      1,
			Cost_id: 1,
			Price:   10000,
			Main:    true,
		},
		{
			ID:      2,
			Cost_id: 3,
			Price:   5000,
			Main:    false,
		},
	}
	if len(request) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}
	return request, nil
}

func (r *Repository) GetCosts() ([]Cost, error) {

	costs := []Cost{ // массив элементов из наших структур
		{
			ID:    1,
			Title: "Аренда офиса",
			Img:   "http://localhost:9000/costs/rent.png",
			Info:  "Аренда офиса относится к постоянным издержкам и не зависит об объема производства. При расчете данной издержки стоит учитывать: стоимость аренды помещения, OPEX (операционные расходы) и т.п. ",
			Type:  true,
		},
		{
			ID:    2,
			Title: "Амортизация",
			Img:   "http://localhost:9000/costs/amortization.png",
			Info:  "Амортизация относится к постоянным издержкам и не зависит от объема производства. При расчете данной издержки стоит учитывать: первоначальную стоимость основного средства, срок его полезного использования и выбранный метод начисления амортизации.",
			Type:  true,
		},
		{
			ID:    3,
			Title: "Стоимость ПО",
			Img:   "http://localhost:9000/costs/PO.png",
			Info:  "Стоимость ПО относится к постоянной издержке и  не зависит от объема производства. При расчете данной издержки стоит учитывать: периодические лицензионные платежи, единовременную стоимость приобретения лицензии, затраты на обновление и техническую поддержку программного обеспечения.",
			Type:  true,
		},
		{
			ID:    4,
			Title: "Расходные материалы",
			Img:   "http://localhost:9000/costs/materials.png",
			Info:  "Расходные материалы относятся к переменным издержкам, так как их потребление напрямую зависит от объема производства. При расчете данных издержек стоит учитывать: стоимость закупки материалов, частоту их использования и норму расхода на единицу продукции.",
			Type:  false,
		},
		{
			ID:    5,
			Title: "Заработная плата",
			Img:   "http://localhost:9000/costs/salary.png",
			Info:  "Заработная плата относится к переменной издержкой и зависит от объема производства. При расчете данных издержек стоит учитывать: сдельная оплата труда производственных рабочих расценки за единицу продукции, объем выпуска, премии за перевыполнение плана. ",
			Type:  false,
		},
		{
			ID:    6,
			Title: "Транспортные расходы",
			Img:   "http://localhost:9000/costs/transport.png",
			Info:  "Транспортные расходы  относится к переменной издержкой и зависит от объема производства. При расчете данных издержек стоит учитывать: стоимость топлива, плата за километраж или тонно-километры при использовании сторонних перевозчиков, разовые затраты на доставку товара.",
			Type:  false,
		},
	}
	if len(costs) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return costs, nil
}

func (r *Repository) GetCost(id int) (Cost, error) {
	// тут у вас будет логика получения нужной услуги, тоже наверное через цикл в первой лабе, и через запрос к БД начиная со второй
	costs, err := r.GetCosts()
	if err != nil {
		return Cost{}, err // тут у нас уже есть кастомная ошибка из нашего метода, поэтому мы можем просто вернуть ее
	}

	for _, cost := range costs {
		if cost.ID == id {
			return cost, nil // если нашли, то просто возвращаем найденный заказ (услугу) без ошибок
		}
	}
	return Cost{}, fmt.Errorf("заказ не найден") // тут нужна кастомная ошибка, чтобы понимать на каком этапе возникла ошибка и что произошло
}

func (r *Repository) GetCostsByTitle(title string) ([]Cost, error) {
	costs, err := r.GetCosts()
	if err != nil {
		return []Cost{}, err
	}

	var result []Cost
	for _, cost := range costs {
		if strings.Contains(strings.ToLower(cost.Title), strings.ToLower(title)) {
			result = append(result, cost)
		}
	}

	return result, nil
}
