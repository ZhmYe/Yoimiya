package utils

import "fmt"

// Utils
// Compute some statistics
// we use variable(...int) to obtain the value of each sample(int)

// Mean
// Compute E(X)
func Mean(variable ...int) int {
	sum := 0
	for _, v := range variable {
		sum += v
	}
	return sum / len(variable)
}

// Variance
// Compute V(X)
func Variance(variable ...int) int {
	// V(X) = E(X^2) - E(X)^2
	// compute E(X^2)
	sum := 0       // to compute E(X)
	squareSum := 0 // to compute E(X^2)
	for _, v := range variable {
		sum += v
		squareSum += v * v
	}
	// here to get higher accuracy
	/***
		如果直接写成V(X) = square / len - (sum / len) ^ 2, 精度会很差，因为分别取整
	 ***/
	return int(float64(squareSum)/float64(len(variable)) -
		(float64(sum)/float64(len(variable)))*(float64(sum)/float64(len(variable))))
}

// SampleVariance
// Compute Sx
func SampleVariance(variable ...int) int {
	// S_x = n * V(X) / (n - 1)
	//v := u.Variance(variable...)
	/***
		这如果直接用上面的方差计算得到结果，精度会很差
	***/
	sum := 0       // to compute E(X)
	squareSum := 0 // to compute E(X^2)
	for _, v := range variable {
		sum += v
		squareSum += v * v
	}
	// n * V(X) / (n - 1) = n * (E(X^2) - E(X)^2) / (n - 1)
	// (n * squareSum - sum * sum) / (n-1)n
	return int(
		float64(len(variable)*squareSum-sum*sum) /
			float64(len(variable)*(len(variable)-1)))
}
func TwoSampleT(X []int, Y []int) int {
	sumX, sumY, squareX, squareY := 0, 0, 0, 0
	for _, x := range X {
		sumX += x
		squareX += x * x
	}
	for _, y := range Y {
		sumY += y
		squareY += y * y
	}
	m1, m2 := len(X), len(Y)
	top := (m2*m2*sumX*sumX + m1*m1*sumY*sumY - 2*m1*m2*sumX*sumY) * (m1 + m2 - 2) * (m1 * m1 * m2 * m2)
	bottom := (m1*m2*m2 + m1*m1*m2) * (m1*m2*m2*squareX - m2*m2*sumX*sumX + m1*m1*m2*squareY - m1*m1*sumY*sumY)
	fmt.Println(top, bottom)
	return int(float64(top) / float64(bottom))
}
