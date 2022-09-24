import {Component, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from "@angular/router";
import {switchMap} from "rxjs";
import {Location} from "@angular/common";

import {FeatureService} from "../services/feature.service";
import {Feature} from "../services/feature";

@Component({
  selector: 'app-feature-detail',
  templateUrl: './feature-detail.component.html',
})
export class FeatureDetailComponent implements OnInit {
  feature: Feature = {
    id: null,
    technicalName: '',
    displayName: null,
    description: null,
    expiresOn: null,
    inverted: false,
    createdAt: 0,
    updatedAt: 0,
  };

  showModal = false;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private location: Location,
    private featureService: FeatureService,
  ) {
  }

  ngOnInit(): void {
    this.route.paramMap.pipe(
      switchMap(params => {
        return this.featureService.getFeature(params.get('featureId')!);
      })
    ).subscribe(feature => {
      this.feature = feature;
    })
  }

  goBack(): void {
    this.location.back();
  }

  startArchive(): void {
    this.showModal = true;
  }

  cancelArchive(): void {
    this.showModal = false;
  }

  archive(): void {
    const that = this;
    this.featureService.archiveFeature(this.feature.id!)
      .subscribe({
        complete() {
          that.router.navigate(['/features']);
        }
      });
  }
}
