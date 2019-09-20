package sort

type Elem interface {
	Less(e Elem) bool
	Equal(e Elem) bool
}

func less(e1, e2 Elem) bool {
	return e1.Less(e2)
}

func greater(e1, e2 Elem) bool {
	return e2.Less(e1)
}

func less_equal(e1, e2 Elem) bool {
	return e1.Less(e2) || e1.Equal(e2)
}

func greater_equal(e1, e2 Elem) bool {
	return !e1.Less(e2)
}

/**
 * quick sort
 */
func qsort_partition(arr []Elem, low, high int) int {
	pivot := arr[low]
	for low < high {
		for low < high && less_equal(pivot, arr[high]) {
			high -= 1
		}
		if low < high {
			arr[low] = arr[high]
		}
		for low < high && less_equal(arr[low], pivot) {
			low += 1
		}
		if low < high {
			arr[high] = arr[low]
		}
	}
	arr[low] = pivot
	return low
}

func QSort(arr []Elem, low, high int) {
	if low < high {
		pivot := qsort_partition(arr, low, high)
		QSort(arr, low, pivot-1)
		QSort(arr, pivot+1, high)
	}
}

/**
 * heap sort
 */
