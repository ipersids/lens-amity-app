import axios from "axios";
import { internalApi } from "./api";

const basePhotoURI = "/api/photos";

export type UploadIntentItem = {
  date: string; // expected format DD-MM-YYYY
  contentType: string;
};

type UploadIntentResponse = {
  photoID: string;
  url: string;
  method: string;
  headers: Record<string, string[]>;
};

const uploadIntent = async (photoInfo: UploadIntentItem): Promise<UploadIntentResponse> => {
  const { data } = await internalApi.put<UploadIntentResponse>(`${basePhotoURI}/upload`, photoInfo);

  return data;
};

const uploadFile = async (intent: UploadIntentResponse, file: File): Promise<void> => {
  const headers: Record<string, string> = {};

  Object.entries(intent.headers).forEach(([key, values]) => {
    if (key.toLowerCase() === "host") return;

    headers[key] = values.join(",");
  });

  if (!headers["Content-Type"]) {
    headers["Content-Type"] = file.type;
  }

  await axios.request({
    url: intent.url,
    method: intent.method,
    data: file,
    headers,
    withCredentials: false,
  });
};

const completeUpload = async (photoID: string): Promise<void> => {
  await internalApi.put(`${basePhotoURI}/${photoID}/complete`);
};

const upload = async (photoInfo: UploadIntentItem, file: File): Promise<void> => {
  const photoIntentData = await uploadIntent(photoInfo);
  await uploadFile(photoIntentData, file);
  await completeUpload(photoIntentData.photoID);
};

const photoService = { upload };

export default photoService;
