import {NgModule} from '@angular/core';
import {CommonModule} from '@angular/common';
import {HttpClientModule} from "@angular/common/http";
import {FormsModule} from "@angular/forms";

import {FeatureRoutingModule} from "./feature-routing.module";
import {FeatureListComponent} from "./feature-list/feature-list.component";
import {NewFeatureComponent} from "./new-feature/new-feature.component";

@NgModule({
  declarations: [
    FeatureListComponent,
    NewFeatureComponent
  ],
  imports: [
    CommonModule,
    FeatureRoutingModule,
    HttpClientModule,
    FormsModule
  ]
})
export class FeatureModule {
}
