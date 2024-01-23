<template>
  <div v-if="this.withSort">
    <label for="attribute">Сортировать по: </label>
    <select
      name="attribute"
      @change="selectChange($event.target.value, this.ascending, this.page_size, this.status, this.title)"
      :value="this.attribute"
    >
      <option value="title">Названию</option>
      <option value="upload_date">Дате загрузки</option>
    </select>
    <select
      name="ascending"
      @change="selectChange(this.attribute, $event.target.value, this.page_size, this.status, this.title)"
      :value="this.ascending"
    >
      <option value="true">По возрастанию</option>
      <option value="false">По убыванию</option>
    </select>
    <br />
    <label for="status">Показать в статусе: </label>
    <select
      name="status"
      @change="
        selectChange(this.attribute, this.ascending, this.page_size, $event.target.value, this.title)
      "
      :value="this.status"
    >
      <option value="Complete">Загружено</option>
      <option value="Archive">Архив</option>
      <option value="Uploading">Загружается</option>
      <option value="Encoding">Обрабатывается</option>
      <option value="Uploaded">Ожидает обработки</option>
      <option value="Fail_upload">Ошибка загрузки</option>
      <option value="Fail_encode">Ошибка обработки</option>
    </select>
    <br />
    <label for="page_size">Количество на странице: </label>
    <select
      name="page_size"
      @change="
        selectChange(this.attribute, this.ascending, $event.target.value, this.status, this.title)
      "
      :value="this.page_size"
    >
      <option value="10">10</option>
      <option value="25">25</option>
      <option value="50">50</option>
      <option value="100">100</option>
    </select>
    <br/>
    <label for="title">Название файла: </label>
    <input name="title" type="text" placeholder="Enter для поиска" @change="
        selectChange(this.attribute, this.ascending, this.page_size, this.status, $event.target.value)
    "/>
  </div>
  <div class="PageNavigation">
    <button
      class="button PageNavigation__button"
      @click="pageChange('first')"
      :disabled="this.is_first"
    >
      <i class="fa-solid fa-backward-fast"></i>
    </button>
    <button
      class="button PageNavigation__button"
      @click="pageChange('previous')"
      :disabled="this.is_first"
    >
      <i class="fa-solid fa-caret-left"></i>
    </button>
    {{ this.page }}
    <button
      class="button PageNavigation__button"
      @click="pageChange('next')"
      :disabled="this.is_last"
    >
      <i class="fa-solid fa-caret-right"></i>
    </button>
    <button
      class="button PageNavigation__button"
      @click="pageChange('last')"
      :disabled="this.is_last"
    >
      <i class="fa-solid fa-forward-fast"></i>
    </button>
  </div>
</template>

<script>
export default {
  name: "PageNavigation",
  props: {
    page: Number,
    page_size: Number,
    is_last: Boolean,
    is_first: Boolean,
    attribute: String,
    ascending: Boolean,
    status: String,
    title: String,
    withSort: Boolean,
  },
  emits: ["selectChange", "pageChange"],
  methods: {
    selectChange: function (attr, asc, pg_size, stat, title) {
      asc = asc == "true";
      this.$emit("selectChange", {
        attribute: attr,
        ascending: asc,
        page_size: pg_size,
        status: stat,
        title: title,
      });
    },
    pageChange: function (i) {
      this.$emit("pageChange", { page: i });
    },
  },
};
</script>

<style scoped lang="scss">
.PageNavigation {
  display: flex;
  flex-direction: row;
  justify-content: center;
  align-items: center;
  column-gap: 1em;
  margin: 1em;
  font-size: 1.5em;

  &__button {
    background-color: rgb(241, 241, 241);
  }
}
</style>
