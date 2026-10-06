import { useInfiniteQuery } from "@tanstack/react-query";
import { Link } from "react-router";
import { ProfilePhotosQueryKey } from "../../../constants";
import type { Photo } from "../../../services/users";
import usersService from "../../../services/users";

type ProfilePhotosProps = {
  username: string;
  canViewPhotos: boolean;
};

const pageSize = 1;

const PhotosFeed = ({ photos, username }: { photos: Photo[]; username: string }) => {
  if (!photos.length) {
    return <p>No photos yet :)</p>;
  }

  return (
    <ul className="profile-photos" aria-label="Profile photos">
      {photos.map((photo) => (
        <li className="profile-photo" key={photo.photoID}>
          <Link
            to={`/users/${username}/photos/${photo.photoID}`}
            aria-label={`${photo.title}, ${photo.date}`}
          >
            <img src={photo.request.url} alt={photo.title} />
            <span className="profile-photo-name">{photo.title}</span>
            <time dateTime={photo.date}>{photo.date}</time>
          </Link>
        </li>
      ))}
    </ul>
  );
};

const ProfilePhotos = ({ username, canViewPhotos }: ProfilePhotosProps) => {
  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isPending, isError } =
    useInfiniteQuery({
      queryKey: ProfilePhotosQueryKey(username),
      queryFn: ({ pageParam }) => usersService.getUserPhotos(username, pageSize, pageParam),
      initialPageParam: undefined as string | undefined,
      getNextPageParam: (lastPage) => lastPage.nextCursor,
      enabled: canViewPhotos && !!username,
      staleTime: 60_000,
    });

  if (!canViewPhotos) {
    return <p>Can't see photos</p>;
  }

  if (isPending) {
    return <p>Loading...</p>;
  }

  if (isError) {
    return <p>Error</p>;
  }

  const photos = data.pages.flatMap((page) => page.items);

  return (
    <>
      <PhotosFeed photos={photos} username={username} />
      {hasNextPage && (
        <button
          className="profile-photos-load-more"
          disabled={isFetchingNextPage}
          type="button"
          onClick={() => fetchNextPage()}
        >
          {isFetchingNextPage ? "Loading..." : "Load more"}
        </button>
      )}
    </>
  );
};

export default ProfilePhotos;
