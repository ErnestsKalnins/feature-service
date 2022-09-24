import {NgModule} from '@angular/core';
import {CommonModule} from '@angular/common';
import {FormsModule} from "@angular/forms";

import {FeatureRoutingModule} from "./feature-routing.module";
import {FeatureListComponent} from "./feature-list/feature-list.component";
import {NewFeatureComponent} from "./new-feature/new-feature.component";
import {EditFeatureComponent} from './edit-feature/edit-feature.component';
import { FeatureDetailComponent } from './feature-detail/feature-detail.component';

@NgModule({
  declarations: [
    FeatureListComponent,
    NewFeatureComponent,
    EditFeatureComponent,
    FeatureDetailComponent
  ],
  imports: [
    CommonModule,
    FeatureRoutingModule,
    FormsModule
  ]
})
export class FeatureModule {
}
