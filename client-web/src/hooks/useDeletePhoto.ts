import { type InfiniteData, useMutation, useQueryClient } from "@tanstack/react-query";
import { PhotoIDQueryKey, ProfilePhotosQueryKey } from "../constants";
import photoService from "../services/photo";
import type { UserPhotosResponse } from "../services/users";
import { useUser } from "../stores/auth";

type PhotosCashe = InfiniteData<UserPhotosResponse, string | undefined>;

const useDeletePhoto = () => {
  const currentUser = useUser();
  const queryClient = useQueryClient();
  const queryKey = ProfilePhotosQueryKey(currentUser?.username ?? "");

  return useMutation({
    mutationFn: photoService.deletePhoto,
    onMutate: async (photoID: string) => {
      await queryClient.cancelQueries({ queryKey });

      const previous = queryClient.getQueryData<PhotosCashe>(queryKey);

      queryClient.setQueryData<PhotosCashe>(queryKey, (current) => {
        if (!current) return current;

        return {
          ...current,
          pages: current.pages.map((page) => ({
            ...page,
            items: page.items.filter((photo) => photo.photoID !== photoID),
          })),
        };
      });

      return { previous };
    },

    onSuccess: (_data, photoID) => {
      if (currentUser) {
        queryClient.invalidateQueries({
          queryKey: PhotoIDQueryKey(currentUser.username, photoID),
          exact: true,
          refetchType: "none",
        });
      }
    },

    onError: (_error, _photoID, context) => {
      if (context?.previous) {
        queryClient.setQueryData(queryKey, context.previous);
      }
    },
  });
};

export default useDeletePhoto;
