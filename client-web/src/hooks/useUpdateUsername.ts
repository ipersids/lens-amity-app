import { type InfiniteData, useMutation, useQueryClient } from "@tanstack/react-query";
import { ProfilePhotosQueryKey, ProfileQueryKey } from "../constants";
import usersService, { type UserPhotosResponse, type UserProfileResponse } from "../services/users";
import { useUpdateUsernameInStore, useUser } from "../stores/auth";

const useUpdateUsername = () => {
  const currentUser = useUser();
  const queryClient = useQueryClient();
  const updateUsernameInStore = useUpdateUsernameInStore();
  const previousProfileKey = ProfileQueryKey(currentUser?.username ?? "");
  const previousProfilePhotosKey = ProfilePhotosQueryKey(currentUser?.username ?? "");

  return useMutation({
    mutationFn: usersService.updateUsername,
    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: previousProfileKey, exact: true });
      await queryClient.cancelQueries({ queryKey: previousProfilePhotosKey, exact: true });

      const previousProfile = queryClient.getQueryData<UserProfileResponse>(previousProfileKey);
      const previousProfilePhotos =
        queryClient.getQueryData<InfiniteData<UserPhotosResponse>>(previousProfilePhotosKey);

      queryClient.removeQueries({
        queryKey: previousProfileKey,
        exact: true,
      });

      queryClient.removeQueries({
        queryKey: previousProfilePhotosKey,
        exact: true,
      });

      return { previousProfile, previousProfilePhotos };
    },

    onError: (_error, _variables, onMutateResult) => {
      if (onMutateResult?.previousProfile) {
        queryClient.setQueryData(previousProfileKey, onMutateResult.previousProfile);
        queryClient.setQueryData(previousProfilePhotosKey, onMutateResult.previousProfilePhotos);
      }
    },

    onSuccess: (result, _variables, onMutateResult) => {
      if (onMutateResult?.previousProfile) {
        const nextProfileKey = ProfileQueryKey(result.updatedUsername);
        queryClient.setQueryData(nextProfileKey, {
          ...onMutateResult.previousProfile,
          username: result.updatedUsername,
        });
      }
      updateUsernameInStore(result.updatedUsername);
    },
  });
};

export default useUpdateUsername;
