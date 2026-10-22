import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ProfilePhotosQueryKey } from "../constants";
import photoService, { type UploadIntentItem } from "../services/photo";
import { useUser } from "../stores/auth";

type AddPhotoVariables = {
  photoInfo: UploadIntentItem;
  file: File;
};

const useAddPhoto = () => {
  const currentUser = useUser();
  const queryClient = useQueryClient();
  const queryKey = ProfilePhotosQueryKey(currentUser?.username ?? "");

  return useMutation({
    mutationFn: ({ photoInfo, file }: AddPhotoVariables) => photoService.upload(photoInfo, file),
    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey });
    },

    onSuccess: () => {
      queryClient.removeQueries({ queryKey, exact: true });
    },
  });
};

export default useAddPhoto;
