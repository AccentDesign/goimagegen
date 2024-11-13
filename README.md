# goimagen

### generating

url in the format of `/images/{operations}/{filename}`

### operations

* blur: `blur=0.5`
* sharpen: `sharpen=0.5`
* gamma: `gamma=0.75`
* contrast: `contrast=20`
* brightness: `brightness=20`
* saturation: `saturation=20`
* resize: `resize=200x0`
* fit: `fit=200x200`
* fill: `fill=200x200@center`
* crop: `crop=200x200@top`
* grayscale: `grayscale`
* invert: `invert`

### multiple operations

eg, fill a 200x200 area from the top of the image, blur and then greyscale:

```bash
http://localhost/images/fill=200x200@top,blur=0.5,grayscale/gorilla.jpeg
```
