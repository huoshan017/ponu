package sort

/*type Elem interface {
	Less(e Elem) bool
	Greater(e Elem) bool
	Equal(e Elem) bool
}*/

func qsort_partition(arr []int, low, high int) int {
	pivot := arr[low]
	for low < high {
		for low < high && arr[high] >= pivot {
			high -= 1
		}
		if low < high {
			arr[low] = arr[high]
		}
		for low < high && arr[low] <= pivot {
			low += 1
		}
		if low < high {
			arr[high] = arr[low]
		}
	}
	arr[low] = pivot
	return low
}

func QSort(arr []int /*Elem*/, low, high int) {
	if low < high {
		pivot := qsort_partition(arr, low, high)
		QSort(arr, low, pivot-1)
		QSort(arr, pivot+1, high)
	}
}
