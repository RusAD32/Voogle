<template>
  <div class="watchview">
    <h1 class="watchview__title">WATCHING</h1>
    <h2 class="watchview__video-title">{{ this.title }} - {{ this.date }}</h2>
    <h2 class="watchview__video-title">
      <a :href="'https://anisurf.site/?id=' + this.id">https://anisurf.site/?id={{ this.id }}</a>
    </h2>
    <VideoPlayer :videoId="this.id" :filterlist="this.filterlist" />
    <FilterSelector @filterListUpdate="updateList" />
    <form class="watchview__form" @submit.prevent="submitFile()">
      <div class="watchview__additional-boxes">
        <div>
          <label class="watchview__form-label" for="videosubs"
          >Выберите субтитры:
          </label>
          <UploadBox
            :title="this.subs.name"
            :accepting="'.ass'"
            :refto="'subtitle_file'"
            @sendFile="handleSubtitles"
          />
        </div>
        <div>
          <label class="watchview__form-label" for="videocover"
          >Выберите обложку:
          </label>
          <UploadBox
            :title="this.cover.name"
            :accepting="'image/jpeg, image/png'"
            :refto="'cover_file'"
            @sendFile="handleCover"
          />
        </div>
      </div>
    <label class="watchview__form-label" for="videotitle"
    >Изменить название видео:
    </label>
    <input
      class="watchview__form-input"
      id="videotitle"
      type="text"
      placeholder="Enter a Title"
      v-model="title"
      required
    />
    <span class="watchview__form-buttoncontainer">
      <button
        type="submit"
        class="button is-primary"
        :disabled="!fileSelected"
      >
        <span>Загрузить</span>
        <span><i class="fa-solid fa-upload"></i></span>
      </button>
      <button
        class="button is-danger is-outlined"
        :disabled="!fileSelected"
        @click.stop.prevent="retry()"
      >
        <span>Отмена</span>
        <span class="icon is-small"> <i class="fa-solid fa-xmark"></i></span>
      </button>
    </span>
    </form>
  </div>
</template>

<script>
import axios from "axios";
import cookies from "js-cookie";
import VideoPlayer from "@/components/VideoPlayer.vue";
import FilterSelector from "@/components/FilterSelector.vue";
import UploadBox from "@/components/UploadBox.vue";

export default {
  name: "VideoPlayerPage.vue",
  data: function () {
    return {
      id: this.$route.params.id,
      title: "",
      cover: "",
      date: "",
      subs: "",
      msg: "",
      filterlist: "",
    };
  },
  computed: {
    fileSelected: function () {
      return !this.subs == "";
    },
  },
  methods: {
    updateList: function (payload) {
      if (payload.filterList.length != 0) {
        this.filterlist = "?filter=";
        this.filterlist += payload.filterList.join("&filter=");
      } else {
        this.filterlist = "";
      }
    },
    handleSubtitles: function (payload) {
      this.subs = payload.file;
    },
    handleCover: function (payload) {
      this.cover = payload.file;
    },
    retry: function () {
      this.title = "";
      this.file = "";
      this.subs = "";
    },
    submitFile: function () {
      // Creating a FormData to POST it as multipart FormData
      const formData = new FormData();
      formData.append("title", this.title);
      formData.append("subs", this.subs);
      formData.append("cover", this.cover);
      axios
        .post(process.env.VUE_APP_API_ADDR + "api/v1/videos/" + this.id + "/edit", formData, {
          headers: {
            "Content-Type": "multipart/form-data",
            Authorization: cookies.get("Authorization"),
          },
        })
        .then(() => {
          this.msg = "Successfully added subtitles";
          this.retry();
        })
        .catch((err) => {
          this.msg = err;
        });
    },
  },
  mounted() {
    axios
      .get(process.env.VUE_APP_API_ADDR + `api/v1/videos/${this.id}/info`, {
        headers: {
          Authorization: cookies.get("Authorization"),
        },
      })
      .then((response) => {
        this.title = response.data["title"];
        this.date = new Date(
          response.data["uploadDateUnix"] * 1000
        ).toLocaleDateString();
      })
      .catch((error) => {
        this.title = error;
      });
  },
  components: {
    VideoPlayer,
    FilterSelector,
    UploadBox,
  },
};
</script>

<style scoped lang="scss">
.watchview {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  row-gap: 20px;

  &__title {
    font-size: 1.5em;
    font-weight: bold;
    padding-top: 1em;
  }

  &__video-title {
    font-size: 1em;
    font-weight: bold;
  }
  &__form {
    padding-top: 1em;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    row-gap: 1em;
    &-label {
      font-size: 1.1em;
    }
    &-input {
      padding: 5px 15px;
    }
  }

  &__additional-boxes {
    display: flex;
    gap: 2rem;
  }
}
</style>
