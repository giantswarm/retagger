package imm

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//
// Code generated by Alibaba Cloud SDK Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

// FramesItem is a nested struct in imm response
type FramesItem struct {
	FacesModifyTime     string          `json:"FacesModifyTime" xml:"FacesModifyTime"`
	OCRModifyTime       string          `json:"OCRModifyTime" xml:"OCRModifyTime"`
	OCRStatus           string          `json:"OCRStatus" xml:"OCRStatus"`
	SourcePosition      string          `json:"SourcePosition" xml:"SourcePosition"`
	Exif                string          `json:"Exif" xml:"Exif"`
	ImageUri            string          `json:"ImageUri" xml:"ImageUri"`
	ImageWidth          int             `json:"ImageWidth" xml:"ImageWidth"`
	ImageFormat         string          `json:"ImageFormat" xml:"ImageFormat"`
	SourceType          string          `json:"SourceType" xml:"SourceType"`
	ModifyTime          string          `json:"ModifyTime" xml:"ModifyTime"`
	FileSize            int             `json:"FileSize" xml:"FileSize"`
	SourceUri           string          `json:"SourceUri" xml:"SourceUri"`
	CreateTime          string          `json:"CreateTime" xml:"CreateTime"`
	FacesStatus         string          `json:"FacesStatus" xml:"FacesStatus"`
	RemarksA            string          `json:"RemarksA" xml:"RemarksA"`
	ImageHeight         int             `json:"ImageHeight" xml:"ImageHeight"`
	RemarksB            string          `json:"RemarksB" xml:"RemarksB"`
	ImageTime           string          `json:"ImageTime" xml:"ImageTime"`
	Orientation         string          `json:"Orientation" xml:"Orientation"`
	Location            string          `json:"Location" xml:"Location"`
	OCRFailReason       string          `json:"OCRFailReason" xml:"OCRFailReason"`
	FacesFailReason     string          `json:"FacesFailReason" xml:"FacesFailReason"`
	TagsFailReason      string          `json:"TagsFailReason" xml:"TagsFailReason"`
	TagsModifyTime      string          `json:"TagsModifyTime" xml:"TagsModifyTime"`
	CelebrityStatus     string          `json:"CelebrityStatus" xml:"CelebrityStatus"`
	CelebrityModifyTime string          `json:"CelebrityModifyTime" xml:"CelebrityModifyTime"`
	CelebrityFailReason string          `json:"CelebrityFailReason" xml:"CelebrityFailReason"`
	TagsStatus          string          `json:"TagsStatus" xml:"TagsStatus"`
	RemarksC            string          `json:"RemarksC" xml:"RemarksC"`
	RemarksD            string          `json:"RemarksD" xml:"RemarksD"`
	ExternalId          string          `json:"ExternalId" xml:"ExternalId"`
	Faces               []FacesItem     `json:"Faces" xml:"Faces"`
	Tags                []TagsItem      `json:"Tags" xml:"Tags"`
	OCR                 []OCRItem       `json:"OCR" xml:"OCR"`
	Celebrity           []CelebrityItem `json:"Celebrity" xml:"Celebrity"`
}
